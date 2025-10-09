package http

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
)

func NewRouter(authHandler AuthHandler) *chi.Mux {
	r := chi.NewRouter()
	logFormat := httplog.SchemaECS.Concise(false)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	})).With(
		slog.String("app", "hris-cmlabs"),
		slog.String("version", "v1.0.0"),
		slog.String("env", "development"),
	)

	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:  slog.LevelDebug,
		Schema: httplog.SchemaECS,
	}))

	r.Use(middleware.AllowContentEncoding("application/json"))
	r.Use(middleware.CleanPath)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/"))

	r.Get("/yo", func(w http.ResponseWriter, r *http.Request) {
		w.Write(([]byte("hello world\n")))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/refresh", func(w http.ResponseWriter, r *http.Request) {})
			r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {})
			r.Post("/forgot-password", func(w http.ResponseWriter, r *http.Request) {})
			r.Post("/verify-email", func(w http.ResponseWriter, r *http.Request) {})

			r.Route("/login", func(r chi.Router) {
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {})
				r.Post("/employee-code", func(w http.ResponseWriter, r *http.Request) {})
				r.Route("/oauth", func(r chi.Router) {
					r.Get("/google", func(w http.ResponseWriter, r *http.Request) {})
				})
			})
		})
	})
	return r
}
