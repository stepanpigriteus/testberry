package main

import (
	"context"
	"fmt"
	"log"
	"testberry/internal/adapters/cache"
	"testberry/internal/adapters/http"
	messagebrok "testberry/internal/adapters/message_brok"
	"testberry/internal/adapters/postgres"
	"testberry/internal/domain/service"
	"testberry/pkg/config"
	"testberry/pkg/logger"
	_  "github.com/lib/pq"
)

func main() {
	logger := logger.NewSlogAdapter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger.Info("[1/4] Reading configurations")
	cfg := config.LoadConfig()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)
	logger.Info("[2/4] Connecting to the DB")
	db, err := postgres.ConnectDB(connStr)
	if err != nil {
		log.Fatalf("could not connect to db: %v", err)
	}
	defer db.Close()
	logger.Info("[3/4] Connecting to Redis")
	cache := cache.NewCache("localhost:6379", "", 0)
	repo := postgres.NewRepository(db)
	broker := messagebrok.NewNoopConsumer()
	service := service.NewService(repo, cache, broker, logger)
	logger.Info("[4/4] Start Server")
	server := http.NewServer(service, ":8081", logger)
	go func() {
		if err := service.Start(ctx); err != nil {
			logger.Error("Service failed", err)
		}
	}()
	if err := server.RunServer(ctx); err != nil {
		logger.Error("HTTP server failed", err)
	}

}
