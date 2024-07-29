package cmd

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"messaggio/broker"
	"messaggio/server"
	"messaggio/storage"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "server",
		Long:  "server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := getConfig()

			log.Trace("server started")
			defer log.Trace("server stopped")

			b, err := broker.New(cfg.Kafka)
			if err != nil {
				log.Fatalf("kafka.New: %s", err)
			}
			defer b.Close(cmd.Context())

			store, err := storage.New(cmd.Context(), cfg.DB)
			if err != nil {
				log.Fatalf("storage.New: %s", err)
			}
			defer store.Close()

			s := server.New(cfg.Server, store, b)
			defer s.Close(cmd.Context())
			s.Start()

			// Ожидание завершения работы
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c
			log.Trace("signal received")
		},
	})
}
