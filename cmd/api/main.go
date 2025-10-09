package main

import (
	"fmt"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/config"
	appHTTP "github.com/cmlabs-hris/hris-backend-go/internal/handler/http"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	serviceAuth "github.com/cmlabs-hris/hris-backend-go/internal/service/auth"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	dsn := cfg.DatabaseURL()
	db, err := database.NewPostgreSQLDB(dsn)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}

	userRepo := postgresql.NewUserRepository(db)
	companyRepo := postgresql.NewCompanyRepository(db)
	JWTRepository := postgresql.NewJWTRepository(db)

	JWTService := jwt.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessExpiration, cfg.JWT.RefreshExpiration)

	authService := serviceAuth.NewAuthService(db, userRepo, companyRepo, JWTService, JWTRepository)

	authHandler := appHTTP.NewAuthHandler(JWTService, authService)

	router := appHTTP.NewRouter(authHandler)

	port := fmt.Sprintf(":%d", cfg.App.Port)
	fmt.Printf("Server running at http://localhost%s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		fmt.Println("Server error:", err)
	}
}
