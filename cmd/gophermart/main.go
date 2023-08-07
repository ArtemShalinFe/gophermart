package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophermart/internal/config"
	"github.com/ArtemShalinFe/gophermart/internal/db"
	"github.com/ArtemShalinFe/gophermart/internal/security"
	"github.com/ArtemShalinFe/gophermart/internal/server"
)

const (
	timeoutServerShutdown = time.Second * 10
	timeoutShutdown       = time.Second * 30
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	// Init logger
	zapl, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialize logger err: %w ", err)
	}
	log := zapl.Sugar()
	defer func(log *zap.SugaredLogger) {
		l := zap.L().Sugar()
		if err := log.Sync(); err != nil {
			fs := "cannot flush buffered log entries err: %w"
			if runtime.GOOS == "darwin" {
				if !errors.Is(err, errors.New("bad file descriptor")) {
					l.Errorf(fs, err)
				}
			} else {
				l.Errorf(fs, err)
			}
		}
		l.Info("flush buffered log entries")
	}(log)

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	componentsErrs := make(chan error, 1)

	// Get config
	cfg := config.GetConfig()
	log.Infof("config %+v", cfg)

	// Init DB
	db, err := db.NewDB(ctx, cfg.DSN, log)
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
	hashc, err := security.NewHashController()
	if err != nil {
		return fmt.Errorf("failed to initialize hashcontroller err: %w", err)
	}

	h, err := server.NewHandlers(cfg.Key, db, log, cfg.TokenExp, hashc)
	if err != nil {
		return fmt.Errorf("failed to initialize handlers err: %w", err)
	}

	// Init and run Server
	srv := server.InitServer(ctx, h, *cfg, log, db)
	go func(errs chan<- error) {
		if err := srv.ListenAndServe(); err != nil {
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
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
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
		log.Fatal("failed to gracefully shutdown the service")
	}()

	return nil
}
