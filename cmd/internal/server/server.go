package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophermart/cmd/internal/adapters"
	"github.com/ArtemShalinFe/gophermart/cmd/internal/config"
	"github.com/ArtemShalinFe/gophermart/cmd/internal/models"
)

type Server struct {
	httpServer         *http.Server
	log                *zap.SugaredLogger
	accIntervalTimeout time.Duration
}

func InitServer(ctx context.Context, h *Handlers, cfg config.Config, log *zap.SugaredLogger, db Storage) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr:    cfg.Address,
			Handler: initRouter(h),
		},
		accIntervalTimeout: time.Duration(cfg.AccrualInterval) * time.Second,
		log:                log,
	}

	a := adapters.NewAccrualClient(cfg, log)
	go s.RunOrderAccruals(ctx, a, db)

	return s
}

func initRouter(h *Handlers) *chi.Mux {
	const orderPath = "/orders"

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(h.RequestLogger)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			h.Register(r.Context(), w, r)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			h.Login(r.Context(), w, r)
		})

		r.Group(func(r chi.Router) {
			r.Use(h.JwtMiddleware)

			r.Post(orderPath, func(w http.ResponseWriter, r *http.Request) {
				h.AddOrder(r.Context(), w, r)
			})

			r.Get("/balance", func(w http.ResponseWriter, r *http.Request) {
				h.GetBalance(r.Context(), w, r)
			})

			r.Post("/balance/withdraw", func(w http.ResponseWriter, r *http.Request) {
				h.AddBalanceWithdrawn(r.Context(), w, r)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(h.JwtMiddleware)
			// Автотесты "Яндекса" не позволяют сжимать результаты этих обработчиков.
			// r.Use(CompressMiddleware)

			r.Get(orderPath, func(w http.ResponseWriter, r *http.Request) {
				h.GetOrders(r.Context(), w, r)
			})

			r.Get("/withdrawals", func(w http.ResponseWriter, r *http.Request) {
				h.GetBalanceMovementHistory(r.Context(), w, r)
			})
		})
	})

	return router
}

func (s *Server) ListenAndServe() error {
	if err := s.httpServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server listen and serve err: %w", err)
		}
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown err: %w", err)
	}
	return nil
}

func (s *Server) RunOrderAccruals(ctx context.Context, a *adapters.Accrual, db Storage) {
	ticker := time.NewTicker(s.accIntervalTimeout)

	errs := make(chan error, 1)
	sleepyChan := make(chan int, 1)

	const defaultOrderChanSize = 10
	orders := make(chan *models.Order, defaultOrderChanSize)

	go func(ctx context.Context, orders chan<- *models.Order, sleepyChan <-chan int, errs chan<- error) {
		for {
			select {
			case <-ctx.Done():
				return
			case timeoutSec := <-sleepyChan:
				time.Sleep(time.Duration(timeoutSec) * time.Second)
			case <-ticker.C:
			}

			ors, err := models.GetOrdersForAccrual(ctx, db)
			if err != nil {
				errs <- fmt.Errorf("failed get orders for accrual err: %w", err)
				continue
			}

			for _, o := range ors {
				orders <- o
			}
		}
	}(ctx, orders, sleepyChan, errs)

	go func(ctx context.Context, orders <-chan *models.Order, sleepyChan chan<- int, errs chan<- error) {
		for o := range orders {
			select {
			case <-ctx.Done():
				return
			default:
			}

			oa, err := a.GetOrderAccrual(ctx, o)
			if err != nil {
				if err.IsOrderNotRegistered() {
					continue
				}
				if err.IsTooManyRequests() {
					timeoutSec, ok := err.TimeoutSec()
					if ok {
						sleepyChan <- timeoutSec
						time.Sleep(time.Duration(timeoutSec) * time.Second)
						continue
					}
				}
				errs <- fmt.Errorf("get order accrual failed err: %w", err)
				continue
			}

			o.Status = oa.Status
			o.Accrual = oa.Accrual

			if err := o.Update(ctx, db); err != nil {
				errs <- fmt.Errorf("update order failed err: %w", err)
			}
		}
	}(ctx, orders, sleepyChan, errs)

	go func(ctx context.Context, errs <-chan error) {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errs:
				s.log.Errorf("failed to run order accruals err: %w", err)
			}
		}
	}(ctx, errs)
}
