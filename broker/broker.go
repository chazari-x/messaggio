package broker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"messaggio/config"
	"messaggio/model"
)

type Broker struct {
	cfg    config.Broker
	writer *kafka.Writer
	reader *kafka.Reader
}

func New(cfg config.Broker) (*Broker, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:     []string{cfg.KafkaAddr},
		Topic:       "messaggio",
		Logger:      log.StandardLogger(),
		ErrorLogger: log.StandardLogger(),
	})
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.KafkaAddr},
		Topic:       "messaggio",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     500 * time.Millisecond,
		Partition:   0,
	})

	log.Trace(cfg.KafkaAddr)

	return &Broker{
		cfg:    cfg,
		writer: writer,
		reader: reader,
	}, nil
}

func (b *Broker) Close() {
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
