package cmd

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"messaggio/broker"
	"messaggio/prometheus"
	"messaggio/receiver"
	"messaggio/sender"
	"messaggio/storage"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "broker",
		Short: "broker",
		Long:  "broker",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := getConfig()

			log.Trace("broker started")
			defer log.Trace("broker stopped")

			p := prometheus.New()

			b, err := broker.New(cfg.Kafka)
			if err != nil {
				log.Fatalf("kafka.New: %s", err)
			}
			defer b.Close(cmd.Context())
			b.Start()

			store, err := storage.New(cmd.Context(), cfg.DB)
			if err != nil {
				log.Fatalf("storage.New: %s", err)
			}
			defer store.Close()

			log.Trace("receiver started")
			r := receiver.New(b, store, p)
			r.Start(cmd.Context())

			log.Trace("sender started")
			s := sender.New(b, store, p)
			s.Start(cmd.Context())

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c
			log.Trace("signal received")

			r.Close()
			log.Trace("receiver stopped")
			s.Close()
			log.Trace("sender stopped")
		},
	})
}
