package config

type Broker struct {
	Http string `envconfig:"BROKER_ADDRESS"`
	Addr string `envconfig:"KAFKA_ADDRESS"`
}

type DataBase struct {
	Addr string `envconfig:"DATABASE_ADDRESS"`
}

type Server struct {
	Http string `envconfig:"SERVER_ADDRESS"`
}

type Grafana struct {
	Addr string `envconfig:"GRAFANA_ADDRESS"`
}

type Prometheus struct {
	Addr string `envconfig:"PROMETHEUS_ADDRESS"`
}
