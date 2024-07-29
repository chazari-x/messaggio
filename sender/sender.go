package sender

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"messaggio/broker"
	"messaggio/model"
	"messaggio/prometheus"
	"messaggio/storage"
)

type Worker struct {
	broker     *broker.Broker
	storage    *storage.Storage
	prometheus *prometheus.Prometheus
	wg         sync.WaitGroup
	closeS     chan struct{}
}

func New(b *broker.Broker, s *storage.Storage, p *prometheus.Prometheus) *Worker {
	return &Worker{
		broker:     b,
		storage:    s,
		prometheus: p,
		closeS:     make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		for {
			select {
			case <-w.closeS:
				log.Info("sender stopped")
				return
			case <-time.After(1 * time.Second):
				msgs, err := w.storage.SelectNew()
				if err != nil {
					log.Error(err)
					continue
				}

				w.prometheus.NewMessageGauge.Add(float64(len(msgs)))

				if len(msgs) == 0 {
					time.Sleep(1 * time.Second)
					continue
				}

				if err = w.broker.Send(ctx, msgs); err != nil {
					log.Error(err)
					continue
				}

				w.prometheus.NewMessageGauge.Sub(float64(len(msgs)))
				w.prometheus.ProcessingMessageGauge.Add(float64(len(msgs)))

				if err = w.storage.UpdateStatuses(msgs, model.Processing); err != nil {
					log.Error(err)
					continue
				}

				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (w *Worker) Close() {
	close(w.closeS)
	w.wg.Wait()
}
