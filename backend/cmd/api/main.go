package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/irham/topup-backend/config"
	"github.com/irham/topup-backend/database"
	"github.com/irham/topup-backend/kafka"
	"github.com/irham/topup-backend/pkg"
	"github.com/irham/topup-backend/router"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pgMain, err := database.NewPostgres(ctx, cfg.DBMain)
	if err != nil {
		log.Fatalf("postgres main: %v", err)
	}
	defer pgMain.Close()
	log.Println("postgres main connected")

	rdb, err := database.NewRedis(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()
	log.Println("redis connected")

	producer := kafka.NewProducer(cfg.Kafka)
	defer producer.Close()
	log.Println("kafka producer ready")

	midtransClient := pkg.NewMidtransClient(cfg.Midtrans)
	log.Printf("midtrans client ready (env=%s)", midtransClient.Env())

	// Suppress "declared and not used" until handlers are wired in.
	_ = pgMain
	_ = rdb
	_ = producer
	_ = midtransClient

	r := router.New(router.Deps{Cfg: cfg})

	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s (env=%s)", cfg.App.Port, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	log.Println("bye")
}
