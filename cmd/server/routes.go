package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/job-crud-app/internal/config"
	"github.com/example/job-crud-app/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

func newRouter(cfg *config.Config, logger *slog.Logger, jh *handlers.Jobs) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.HTTPRequestTimeout))
	r.Use(accessLogMiddleware(logger))

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"service": cfg.ServiceName}); err != nil {
			logger.Error("write service info response", slog.Any("err", err))
		}
	})

	r.Route("/api/jobs", func(api chi.Router) {
		api.Get("/", jh.List)
		api.Post("/", jh.Create)
		api.Get("/{id}", jh.Get)
		api.Delete("/{id}", jh.Delete)
	})

	return r
}

func accessLogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
