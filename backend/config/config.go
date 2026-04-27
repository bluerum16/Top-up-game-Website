package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	App      AppConfig
	DBMain   PostgresConfig
	DBDash   PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	Midtrans MidtransConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Env  string
	Port string
}

type PostgresConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Pass, p.Host, p.Port, p.Name,
	)
}

type RedisConfig struct {
	Host string
	Port string
	Pass string
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type KafkaConfig struct {
	Brokers []string
}

type MidtransConfig struct {
	Env       string
	ServerKey string
	ClientKey string
}

func (m MidtransConfig) IsProduction() bool {
	return strings.EqualFold(m.Env, "production")
}

type JWTConfig struct {
	Secret string
}

func Load() Config {
	return Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
		DBMain: PostgresConfig{
			Host: os.Getenv("DB_MAIN_HOST"),
			Port: getEnv("DB_MAIN_PORT", "5432"),
			User: os.Getenv("DB_MAIN_USER"),
			Pass: os.Getenv("DB_MAIN_PASS"),
			Name: os.Getenv("DB_MAIN_NAME"),
		},
		DBDash: PostgresConfig{
			Host: os.Getenv("DB_DASH_HOST"),
			Port: getEnv("DB_DASH_PORT", "5432"),
			User: os.Getenv("DB_DASH_USER"),
			Pass: os.Getenv("DB_DASH_PASS"),
			Name: os.Getenv("DB_DASH_NAME"),
		},
		Redis: RedisConfig{
			Host: os.Getenv("REDIS_HOST"),
			Port: getEnv("REDIS_PORT", "6379"),
			Pass: os.Getenv("REDIS_PASS"),
		},
		Kafka: KafkaConfig{
			Brokers: splitCSV(getEnv("KAFKA_BROKERS", "kafka:9092")),
		},
		Midtrans: MidtransConfig{
			Env:       getEnv("MIDTRANS_ENV", "sandbox"),
			ServerKey: os.Getenv("MIDTRANS_SERVER_KEY"),
			ClientKey: os.Getenv("MIDTRANS_CLIENT_KEY"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
