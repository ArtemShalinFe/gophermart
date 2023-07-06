package server

import (
	"net/http"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/config"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func InitServer(h *Handlers, cfg *config.Config) *http.Server {
	return &http.Server{
		Addr:    cfg.Address,
		Handler: initRouter(h),
	}
}

func initRouter(h *Handlers) *chi.Mux {

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		h.Ping(w)
	})

	return router

}
