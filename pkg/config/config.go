package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DB struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Redis struct {
		Host         string        `env:"REDIS_HOST"`
		Port         int           `env:"REDIS_PORT"`
		Password     string        `env:"REDIS_PASSWORD"`
		DB           int           `env:"REDIS_DB"`
		DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT"`
		ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT"`
		WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT"`
		PoolSize     int           `env:"REDIS_POOL_SIZE"`
		TLS          bool          `env:"REDIS_TLS"`
	}
	Kafka struct {
		Brokers       []string `env:"KAFKA_BROKERS"`
		Topic         string   `env:"KAFKA_TOPIC"`
		ConsumerGroup string   `env:"KAFKA_CONSUMER_GROUP"`
	}
}

func LoadConfig() *Config {
	cfg := &Config{}

	cfg.DB.Host = getEnv("DB_HOST")
	cfg.DB.Port = getEnv("DB_PORT")
	cfg.DB.User = getEnv("DB_USER")
	cfg.DB.Password = getEnv("DB_PASSWORD")
	cfg.DB.Name = getEnv("DB_NAME")
	cfg.DB.SSLMode = getEnv("DB_SSLMODE")

	cfg.Redis.Host = getEnv("REDIS_HOST")
	cfg.Redis.Port = mustAtoi("REDIS_PORT", 6379)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD")
	cfg.Redis.DB = mustAtoi("REDIS_DB", 0)
	cfg.Redis.DialTimeout = mustParseDuration("REDIS_DIAL_TIMEOUT", 5*time.Second)
	cfg.Redis.ReadTimeout = mustParseDuration("REDIS_READ_TIMEOUT", 3*time.Second)
	cfg.Redis.WriteTimeout = mustParseDuration("REDIS_WRITE_TIMEOUT", 3*time.Second)
	cfg.Redis.PoolSize = mustAtoi("REDIS_POOL_SIZE", 10)
	cfg.Redis.TLS = mustParseBool("REDIS_TLS", false)

	cfg.Kafka.Brokers = mustParseStringSlice("KAFKA_BROKERS", []string{"localhost:9092"})
	cfg.Kafka.Topic = getEnvWithDefault("KAFKA_TOPIC", "orders")
	cfg.Kafka.ConsumerGroup = getEnvWithDefault("KAFKA_CONSUMER_GROUP", "my-consumer-group")

	return cfg
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("Warning: environment variable %s not set", key)
	}
	return value
}

func getEnvWithDefault(key, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("Warning: %s not set, using default %s", key, defaultVal)
		return defaultVal
	}
	return value
}

func mustAtoi(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Printf("Warning: %s not set, using default %d", key, defaultVal)
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Invalid int for %s: %s, using default %d", key, valStr, defaultVal)
		return defaultVal
	}
	return val
}

func mustParseBool(key string, defaultVal bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Printf("Warning: %s not set, using default %v", key, defaultVal)
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		log.Printf("Invalid bool for %s: %s, using default %v", key, valStr, defaultVal)
		return defaultVal
	}
	return val
}

func mustParseDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Printf("Warning: %s not set, using default %s", key, defaultVal)
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		log.Printf("Invalid duration for %s: %s, using default %s", key, valStr, defaultVal)
		return defaultVal
	}
	return val
}

func mustParseStringSlice(key string, defaultVal []string) []string {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Printf("Warning: %s not set, using default %v", key, defaultVal)
		return defaultVal
	}

	parts := strings.Split(valStr, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		log.Printf("Warning: %s is empty, using default %v", key, defaultVal)
		return defaultVal
	}

	return result
}
