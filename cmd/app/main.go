package main

import (
	"context"
	"go-pet-shop/internal/config"
	"go-pet-shop/internal/handlers"
	"go-pet-shop/internal/handlers/shop"
	"go-pet-shop/internal/lib/logger"
	"go-pet-shop/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DatabaseTimeout)
	defer cancel()

	store, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer store.Close()

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(logger.CustomLogger(log))

	h := shop.New(log, store)
	router.Get("/status", handlers.StatusHandler)
	router.Post("/users", h.CreateUser)
	router.Get("/users", h.GetAllUsers)
	router.Get("/users/{email}", h.GetUserByEmail)

	router.Post("/products", h.CreateProduct)
	router.Get("/products", h.GetAllProducts)
	router.Get("/products/{id}", h.GetProductByID)
	router.Put("/products/{id}", h.UpdateProduct)
	router.Delete("/products/{id}", h.DeleteProduct)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", slog.String("error", err.Error()))
	}
}
