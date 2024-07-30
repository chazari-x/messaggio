package sender

import (
	"context"
	"net/http"
	"strconv"
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
	closeS  chan struct{}
	cfg     config.Broker
}

func New(b *broker.Broker, s *storage.Storage, cfg config.Broker) *Worker {
	return &Worker{
		broker:  b,
		storage: s,
		closeS:  make(chan struct{}),
		cfg:     cfg,
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

				if len(msgs) == 0 {
					time.Sleep(1 * time.Second)
					continue
				}

				if err = w.broker.Send(ctx, msgs); err != nil {
					log.Error(err)
					continue
				}

				client := http.Client{}
				request, err := http.NewRequest("PUT", "http://"+w.cfg.ServerAddr+"/api/messages/processing/add/"+strconv.Itoa(len(msgs)), nil)
				if err == nil {
					if _, err = client.Do(request); err != nil {
						log.Error(err)
					}
				} else {
					log.Error(err)
				}

				request, err = http.NewRequest("PUT", "http://"+w.cfg.ServerAddr+"/api/messages/new/sub/"+strconv.Itoa(len(msgs)), nil)
				if err == nil {
					if _, err = client.Do(request); err != nil {
						log.Error(err)
					}
				} else {
					log.Error(err)
				}

				if err = w.storage.UpdateStatuses(msgs, model.Processing); err != nil {
					log.Error(err)
					continue
				}
			}
		}
	}()
}

func (w *Worker) Close() {
	close(w.closeS)
	w.wg.Wait()
}
