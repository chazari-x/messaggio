package receiver

import (
	"context"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"messaggio/broker"
	"messaggio/config"
	"messaggio/model"
	"messaggio/storage"
)

type Worker struct {
	broker  *broker.Broker
	storage *storage.Storage
	wg      sync.WaitGroup
	closeR  chan struct{}
	cfg     config.Broker
}

func New(b *broker.Broker, s *storage.Storage, cfg config.Broker) *Worker {
	return &Worker{
		broker:  b,
		storage: s,
		closeR:  make(chan struct{}),
		cfg:     cfg,
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
			case <-time.After(100 * time.Millisecond):
				msg, err := w.broker.Recv(ctx)
				if err != nil {
					log.Error(err)
					continue
				}

				client := http.Client{}
				request, err := http.NewRequest("PUT", "http://"+w.cfg.ServerAddr+"/api/messages/processing/sub/1", nil)
				if err == nil {
					if _, err = client.Do(request); err != nil {
						log.Error(err)
					}
				} else {
					log.Error(err)
				}

				request, err = http.NewRequest("PUT", "http://"+w.cfg.ServerAddr+"/api/messages/ok/add/1", nil)
				if err == nil {
					if _, err = client.Do(request); err != nil {
						log.Error(err)
					}
				} else {
					log.Error(err)
				}

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
