package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/config"
)

type Server struct {
	HTTPServer *http.Server
	Log        *zap.SugaredLogger
}

func InitServer(h *Handlers, cfg *config.Config, log *zap.SugaredLogger) *Server {
	return &Server{
		HTTPServer: &http.Server{
			Addr:    cfg.Address,
			Handler: initRouter(h),
		},
		Log: log,
	}
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
				h.AddBalanceWithdraw(r.Context(), w, r)
			})

			r.Get("/withdrawals", func(w http.ResponseWriter, r *http.Request) {
				h.GetWithdrawals(r.Context(), w, r)
			})
		})
	})
	return router
}

func (s *Server) RunOrderAccruals(ctx context.Context) error {
	return nil
}
