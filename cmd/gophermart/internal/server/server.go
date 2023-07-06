package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func InitServer(h *Handlers) *http.Server {
	return &http.Server{
		Addr:    ":8080",
		Handler: initRouter(h),
	}
}

func initRouter(h *Handlers) *chi.Mux {

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {

		// rctx := r.Context()

		// handlers.Ping(rctx, w)

	})

	return router

}
