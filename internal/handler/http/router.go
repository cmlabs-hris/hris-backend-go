package http

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/middleware"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/jwtauth/v5"
)

func NewRouter(JWTService jwt.Service, authHandler AuthHandler, companyhandler CompanyHandler) *chi.Mux {
	r := chi.NewRouter()
	logFormat := httplog.SchemaECS.Concise(false)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	})).With(
		slog.String("app", "hris-cmlabs"),
		slog.String("version", "v1.0.0"),
		slog.String("env", "development"),
	)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		MaxAge:           300,
	}))

	// r.Use(chiMiddleware.RealIP)

	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:  slog.LevelDebug,
		Schema: httplog.SchemaECS,
	}))

	r.Use(chiMiddleware.AllowContentEncoding("application/json"))
	r.Use(chiMiddleware.CleanPath)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Heartbeat("/"))

	r.Get("/yo", func(w http.ResponseWriter, r *http.Request) {
		w.Write(([]byte("hello world\n")))
	})

	r.Route("/api/v1", func(r chi.Router) {

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/logout", authHandler.Logout)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Route("/oauth/callback", func(r chi.Router) {
				r.Get("/google", authHandler.OAuthCallbackGoogle)
			})

			r.Route("/login", func(r chi.Router) {
				r.Post("/", authHandler.Login)
				r.Post("/employee-code", authHandler.LoginWithEmployeeCode)
				r.Route("/oauth", func(r chi.Router) {
					r.Get("/google", authHandler.LoginWithGoogle)
				})
			})

		})

		// Requires authentication
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(JWTService.JWTAuth()))
			r.Use(middleware.AuthRequired(JWTService.JWTAuth()))

			r.Route("/companies", func(r chi.Router) {

				// Admin only
				r.Group(func(r chi.Router) {
					r.Use(middleware.AdminOnly)
					r.Get("/", companyhandler.List)
					r.Post("/", companyhandler.Create)
				})

				r.Route("/my", func(r chi.Router) {
					r.Get("/", companyhandler.GetByID)

					// Admin only
					r.Group(func(r chi.Router) {
						r.Use(middleware.AdminOnly)
						r.Put("/", companyhandler.Update)
						r.Delete("/", companyhandler.Delete)
					})
				})
			})
		})
	})
	return r
}
