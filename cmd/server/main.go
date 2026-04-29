package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gerry1818/wallet-transfer-assignment/internal/db"
	"github.com/gerry1818/wallet-transfer-assignment/internal/handler"
	"github.com/gerry1818/wallet-transfer-assignment/internal/logger"
	"github.com/gerry1818/wallet-transfer-assignment/internal/metrics"
	"github.com/gerry1818/wallet-transfer-assignment/internal/repository/postgres"
	"github.com/gerry1818/wallet-transfer-assignment/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// ✅ Init logger
	logger.Init()
	defer logger.Log.Sync()

	logger.Log.Info("starting wallet service...")

	// ✅ Init metrics
	metrics.Init()

	dbURL := "postgres://postgres:postgres@localhost:5432/wallet?sslmode=disable"

	// ✅ Create DB pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(dbURL)
	if err != nil {
		logger.Log.Fatal("failed to connect to DB", zap.Error(err))
	}

	// Optional: ping DB
	if err := pool.Ping(ctx); err != nil {
		logger.Log.Fatal("database not reachable", zap.Error(err))
	}

	logger.Log.Info("connected to database")

	// ✅ Wire dependencies
	repo := postgres.NewRepo(pool)
	svc := service.NewTransferService(repo)
	h := handler.NewHandler(svc)

	// ✅ Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/transfers", h.Transfer)
	mux.Handle("/metrics", promhttp.Handler())

	// ✅ HTTP server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	// ✅ Start server in goroutine
	go func() {
		logger.Log.Info("server started", zap.String("port", "8080"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("server failed", zap.Error(err))
		}
	}()

	// ✅ Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Log.Info("shutting down server...", zap.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("graceful shutdown failed", zap.Error(err))
	} else {
		logger.Log.Info("server exited cleanly")
	}
}
