package config

type Broker struct {
	KafkaAddr  string
	ServerAddr string
}

type DataBase struct {
	Addr string
}

type Server struct {
	Http           string
	PrometheusAddr string
}
