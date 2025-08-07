package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testberry/internal/adapters/cache"
	"testberry/internal/adapters/http"
	messagebrok "testberry/internal/adapters/message_brok"
	"testberry/internal/adapters/postgres"
	"testberry/internal/domain/service"
	"testberry/pkg/config"
	"testberry/pkg/logger"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	logger := logger.NewSlogAdapter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("[1/7] Reading configurations")
	cfg := config.LoadConfig()

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	logger.Info("[2/7] Connecting to the DB")
	db, err := postgres.ConnectDB(connStr)
	if err != nil {
		log.Fatalf("could not connect to db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close db", err)
		}
	}()

	logger.Info("[3/7] Connecting to Redis")
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	cacheClient := cache.NewCache(redisAddr, cfg.Redis.Password, cfg.Redis.DB)
	repo := postgres.NewRepository(db, logger)

	logger.Info("[4/7] Create Kafka Consumer")
	kafkaBrokers := cfg.Kafka.Brokers
	kafkaTopic := cfg.Kafka.Topic

	consumer, err := messagebrok.NewConsumer(kafkaBrokers, "order-consumer-group", kafkaTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error("failed to close kafka consumer", err)
		}
	}()

	logger.Info("[5/7] Create Kafka Producer")
	producer, err := messagebrok.NewProducer(kafkaBrokers, kafkaTopic)
	if err != nil {
		log.Fatalf("Failed to start Kafka producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			logger.Error("failed to close kafka producer", err)
		}
	}()

	service := service.NewService(repo, cacheClient, consumer, producer, logger)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("[6/7] Starting HTTP Server")
		server := http.NewServer(service, ":8081", logger)
		if err := server.RunServer(ctx); err != nil {
			logger.Error("HTTP server failed", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("[7/7] Restoring Cache")
		if err := service.Start(ctx); err != nil {
			logger.Error("RestoreCacheService failed", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("[8/8] Starting Kafka Consumer")
		if err := service.SaveOrder(ctx); err != nil {
			logger.Error("Kafka consumer service failed", err)
		}
	}()

	logger.Info("Starting Kafka Producer")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	logger.Info("All services started successfully")

	for {
		select {
		case <-ticker.C:
			if err := service.SendRandomOrder(ctx); err != nil {
				logger.Error("Failed to send random order:", err)
			}

		case sig := <-sigChan:
			logger.Info(fmt.Sprintf("Received signal %s, shutting down...", sig))
			cancel()
			ticker.Stop()
			wg.Wait()
			logger.Info("Application shutdown complete")
			return

		case <-ctx.Done():
			logger.Info("Context cancelled, shutting down...")
			ticker.Stop()
			wg.Wait()
			return
		}
	}
}
