package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	DB_URL string
}

func Load() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Faild to Load Config file!!!")
	}

	cfg := &Config{
		Port:   getEnv("Port", "8080"),
		DB_URL: getEnv("DB_URL", ""),
	}

	if cfg.DB_URL == "" {
		log.Fatal("Faild to Load DB Path")
	}

	return cfg

}

func getEnv(Key string, fallback string) string {
	if v := os.Getenv(Key); v != "" {
		return v
	}
	return fallback
}
