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

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/adapters"
	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/config"
	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/models"
)

type Server struct {
	HTTPServer *http.Server
	Log        *zap.SugaredLogger
}

func InitServer(ctx context.Context, h *Handlers, cfg *config.Config, log *zap.SugaredLogger, db Storage) *Server {
	s := &Server{
		HTTPServer: &http.Server{
			Addr:    cfg.Address,
			Handler: initRouter(h),
		},
		Log: log,
	}

	a := adapters.NewAccrualClient(cfg)
	go s.RunOrderAccruals(ctx, a, db)

	return s
}

func initRouter(h *Handlers) *chi.Mux {
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

			r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
				h.AddOrder(r.Context(), w, r)
			})

			r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {
				h.GetOrders(r.Context(), w, r)
			})

			r.Get("/balance", func(w http.ResponseWriter, r *http.Request) {
				h.GetBalance(r.Context(), w, r)
			})

			r.Post("/balance/withdraw", func(w http.ResponseWriter, r *http.Request) {
				h.AddBalanceWithdrawn(r.Context(), w, r)
			})

			r.Get("/withdrawals", func(w http.ResponseWriter, r *http.Request) {
				h.GetBalanceMovementHistory(r.Context(), w, r)
			})
		})
	})
	return router
}

func (s *Server) RunOrderAccruals(ctx context.Context, a models.AccrualService, db Storage) {
	ticker := time.NewTicker(5 * time.Second)

	errs := make(chan error, 1)

	const defaultOrderChanSize = 10
	orders := make(chan *models.Order, defaultOrderChanSize)

	go func(ctx context.Context, orders chan<- *models.Order, errs chan<- error) {
		for {
			ors, err := models.GetOrdersForAccrual(ctx, db)
			if err != nil {
				errs <- fmt.Errorf("failed get orders for accrual err: %w", err)
				continue
			}

			for _, o := range ors {
				orders <- o
			}

			select {
			case <-ctx.Done():
			case <-ticker.C:
			}
		}
	}(ctx, orders, errs)

	go func(ctx context.Context, orders <-chan *models.Order, errs chan<- error) {
		for o := range orders {
			a, err := a.GetOrderAccrual(ctx, o)
			if err != nil {
				if !errors.Is(err, adapters.ErrOrderNotRegistered) {
					errs <- fmt.Errorf("get order accrual failed err: %w", err)
				}
				continue
			}

			o.Status = a.Status
			o.Accrual = a.Accrual

			if err := o.Update(ctx, db); err != nil {
				errs <- fmt.Errorf("update order failed err: %w", err)
			}
		}
	}(ctx, orders, errs)

	go func(ctx context.Context, errs <-chan error) {
		for {
			select {
			case <-ctx.Done():
			case err := <-errs:
				s.Log.Errorf("failed to run order accruals err: %w", err)
			}
		}
	}(ctx, errs)

}
