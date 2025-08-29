package config

import "os"

type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string
	PostgresDSN  string
	HTTPPort     string
}

func Load() *Config {
	return &Config{
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
		KafkaGroupID: os.Getenv("KAFKA_GROUP_ID"),
		PostgresDSN:  os.Getenv("POSTGRES_DSN"),
		HTTPPort:     os.Getenv("HTTP_PORT"),
	}
}
