package receiver

import (
	"context"
	"sync"

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
	closeR     chan struct{}
}

func New(b *broker.Broker, s *storage.Storage, p *prometheus.Prometheus) *Worker {
	return &Worker{
		broker:     b,
		storage:    s,
		prometheus: p,
		closeR:     make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		for {
			select {
			case <-w.closeR:
				log.Info("receiver stopped")
				return
			default:
				msg, err := w.broker.Recv(ctx)
				if err != nil {
					log.Error(err)
					continue
				}

				w.prometheus.ProcessingMessageGauge.Dec()
				w.prometheus.OkMessageCounter.Inc()

				if err = w.storage.UpdateStatus(msg.ID, model.Ok); err != nil {
					log.Error(err)
					continue
				}
			}
		}
	}()
}

func (w *Worker) Close() {
	close(w.closeR)
	w.wg.Wait()
}
