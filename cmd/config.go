package cmd

import (
	"fmt"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"
	"messaggio/config"
)

type Config struct {
	DB         config.DataBase
	Server     config.Server
	Kafka      config.Broker
	Grafana    config.Grafana
	Prometheus config.Prometheus
}

func getConfig() *Config {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			return "", fmt.Sprintf("%s:%d", frame.Function, frame.Line)
		},
	})

	log.SetReportCaller(true)

	log.SetLevel(log.TraceLevel)

	return &Config{
		Server: config.Server{
			Http: os.Getenv("SERVER_ADDRESS"),
		},
		DB: config.DataBase{
			Addr: os.Getenv("DATABASE_ADDRESS"),
		},
		Kafka: config.Broker{
			Http: os.Getenv("BROKER_ADDRESS"),
			Addr: os.Getenv("KAFKA_ADDRESS"),
		},
		Grafana: config.Grafana{
			Addr: os.Getenv("GRAFANA_ADDRESS"),
		},
		Prometheus: config.Prometheus{
			Addr: os.Getenv("PROMETHEUS_ADDRESS"),
		},
	}
}
