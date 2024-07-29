package broker

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"messaggio/config"
	"messaggio/model"
)

type Broker struct {
	cfg    config.Broker
	writer *kafka.Writer
	reader *kafka.Reader
	server *http.Server
}

func New(cfg config.Broker) (*Broker, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.Addr},
		Topic:   "messaggio",
	})
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.Addr},
		Topic:       "messaggio",
		GroupID:     "messaggio-group",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
	})

	return &Broker{
		cfg:    cfg,
		writer: writer,
		reader: reader,
	}, nil
}

func (b *Broker) Start() {
	go func() {
		r := chi.NewRouter()
		r.Handle("/metrics", promhttp.Handler())

		b.server = &http.Server{
			Addr:    b.cfg.Http,
			Handler: r,
		}

		log.Printf("http server starting on %s", b.cfg.Http)
		if err := b.server.ListenAndServe(); err != nil {
			log.Printf("failed to start server: %v", err)
		}
		log.Printf("http server stopped")
	}()
}

func (b *Broker) Close(ctx context.Context) {
	if b.server != nil {
		_ = b.server.Shutdown(ctx)
	}
	_ = b.writer.Close()
	_ = b.reader.Close()
}

func (b *Broker) Send(ctx context.Context, msg []model.Message) error {
	var messages []kafka.Message
	for _, m := range msg {
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		messages = append(messages, kafka.Message{
			Value: data,
		})
	}

	return b.writer.WriteMessages(ctx, messages...)
}

func (b *Broker) Recv(ctx context.Context) (model.Message, error) {
	var msg model.Message
	for {
		m, err := b.reader.ReadMessage(ctx)
		if err != nil {
			if kafkaError, ok := err.(kafka.Error); ok && kafkaError.Temporary() {
				log.Printf("Temporary error while fetching message: %v, retrying...", err)
				time.Sleep(1 * time.Second)
				continue
			}
			return msg, err
		}

		err = json.Unmarshal(m.Value, &msg)
		if err != nil {
			return msg, err
		}
		return msg, nil
	}
}
