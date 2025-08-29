package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"web_service/internal/config"
	"web_service/internal/delivery/http_handler"
	"web_service/internal/delivery/kafka_listener"
	"web_service/internal/infrastructure/cache"
	"web_service/internal/infrastructure/persistent"
	"web_service/internal/infrastructure/persistent/repositories"
	"web_service/internal/usecase"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	cfg := config.Load()

	pool, err := initDatabase(ctx, cfg)
	if err != nil {
		log.Fatal("Database initialization failed: ", err)
	}
	defer pool.Close()

	getOrderUseCase, saveOrderUseCase := initUseCases(pool)

	err = getOrderUseCase.RestoreCache(ctx)
	if err != nil {
		log.Fatal("Cache restore failed: ", err)
	}

	kafkaConsumer, err := startKafkaConsumer(cfg, saveOrderUseCase, ctx)
	if err != nil {
		log.Fatal("Kafka consumer startup failed: ", err)
	}

	server := startHTTPServer(cfg, getOrderUseCase)

	waitForShutdown(server, kafkaConsumer, pool)
}

func initDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}
	dbConfig.MaxConns = 10
	dbConfig.MinConns = 2
	dbConfig.MaxConnLifetime = time.Hour
	dbConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	log.Println("Database connected successfully")
	return pool, nil
}

func initUseCases(pool *pgxpool.Pool) (*usecase.GetOrderUseCase, *usecase.SaveOrderUseCase) {
	transactionManager := persistent.NewPgxTransactionManager(pool)
	orderRepository := repositories.NewOrderRepo(pool)
	paymentRepository := repositories.NewPaymentRepo(pool)
	deliveryRepository := repositories.NewDeliveryRepo(pool)
	itemRepository := repositories.NewItemRepo(pool)

	orderStorage := cache.NewLocalOrderStorage()

	getOrderUseCase := usecase.NewGetOrderUseCase(
		orderRepository, paymentRepository, deliveryRepository, itemRepository, transactionManager, orderStorage)
	saveOrderUseCase := usecase.NewSaveOrderUseCase(
		orderRepository, paymentRepository, deliveryRepository, itemRepository, transactionManager)

	return getOrderUseCase, saveOrderUseCase
}

func startKafkaConsumer(cfg *config.Config,
	saveOrderUseCase *usecase.SaveOrderUseCase, ctx context.Context) (*kafka_listener.Consumer, error) {
	kafkaConsumer, err := kafka_listener.NewConsumer(cfg.KafkaBrokers, cfg.KafkaGroupID, saveOrderUseCase)
	if err != nil {
		return nil, err
	}
	err = kafkaConsumer.Subscribe(cfg.KafkaTopic)
	if err != nil {
		return nil, err
	}
	go func() {
		log.Printf("Starting Kafka consumer for topic: %s", cfg.KafkaTopic)
		kafkaConsumer.Consume(ctx)
	}()

	return kafkaConsumer, nil
}

func startHTTPServer(cfg *config.Config, getOrderUseCase *usecase.GetOrderUseCase) *http.Server {
	handler := http_handler.NewOrderHandler(getOrderUseCase)

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP server failed: ", err)
		}
	}()

	return server
}

func waitForShutdown(server *http.Server, kafkaConsumer *kafka_listener.Consumer, pool *pgxpool.Pool) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down gracefully...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Println("HTTP server stopped gracefully")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if kafkaConsumer != nil {
			kafkaConsumer.Close()
			log.Println("Kafka consumer stopped gracefully")
		}
	}()

	wg.Wait()

	if pool != nil {
		pool.Close()
		log.Println("Database connection closed")
	}

	log.Println("Application shutdown completed")
}
