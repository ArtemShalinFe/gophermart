package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/config"
	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/db"
	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/server"
)

const (
	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	// Init logger
	zapl, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialize logger err: %w ", err)
	}
	log := zapl.Sugar()

	componentsErrs := make(chan error, 1)

	wg.Add(1)
	go func(errs chan<- error) {
		defer log.Info("flush buffered log entries")
		defer wg.Done()
		<-ctx.Done()

		if err := log.Sync(); err != nil {
			if runtime.GOOS != "darwin" {
				errs <- fmt.Errorf("cannot flush buffered log entries err: %w", err)
			}
		}
	}(componentsErrs)

	// Get config
	cfg := config.GetConfig()
	log.Infof("config %+v", cfg)

	// Init DB
	db, err := db.NewDB(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to initialize DB err: %w", err)
	}

	wg.Add(1)
	go func() {
		defer log.Info("closed DB")
		defer wg.Done()
		<-ctx.Done()

		db.Close()
	}()

	// Init Handlers
	h, err := server.NewHandlers(cfg.Key, db, log)
	if err != nil {
		return fmt.Errorf("failed to initialize handlers err: %w", err)
	}

	// Init and run Server
	srv := server.InitServer(ctx, h, cfg, log, db)
	go func(errs chan<- error) {
		if err := srv.HTTPServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			errs <- fmt.Errorf("listen and server has failed: %w", err)
		}
	}(componentsErrs)

	// Graceful shutdown
	wg.Add(1)
	go func() {
		defer log.Error("server has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err := srv.HTTPServer.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Errorf("an error occurred during server shutdown: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		log.Error(err)
		cancelCtx()
	}

	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Error("failed to gracefully shutdown the service")
	}()

	return nil
}
