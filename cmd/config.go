package cmd

import (
	"fmt"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"
	"messaggio/config"
)

type Config struct {
	DB     config.DataBase
	Server config.Server
	Kafka  config.Broker
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
			Http:           os.Getenv("SERVER_HTTP"),
			PrometheusAddr: os.Getenv("PROMETHEUS_ADDRESS"),
		},
		DB: config.DataBase{
			Addr: os.Getenv("DATABASE_ADDRESS"),
		},
		Kafka: config.Broker{
			KafkaAddr:  os.Getenv("KAFKA_ADDRESS"),
			ServerAddr: os.Getenv("SERVER_ADDRESS"),
		},
	}
}
