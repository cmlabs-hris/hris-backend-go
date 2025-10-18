package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database     DatabaseConfig
	JWT          JWTConfig
	App          AppConfig
	OAuth2Google OAuth2GoogleConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret            string
	RefreshExpiration string
	AccessExpiration  string
}

// AppConfig holds application configuration
type AppConfig struct {
	Port     int
	Env      string
	LogLevel string
}

type OAuth2GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := &Config{}

	// Database configuration
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	config.Database = DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		Name:     getEnv("DB_NAME", "cmlasb-hris"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	// Redis configuration
	// redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	// }

	// redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	// }

	// config.Redis = RedisConfig{
	// 	Host:     getEnv("REDIS_HOST", "localhost"),
	// 	Port:     redisPort,
	// 	Password: getEnv("REDIS_PASSWORD", ""),
	// 	DB:       redisDB,
	// }

	// Application configuration
	appPort, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %w", err)
	}

	config.App = AppConfig{
		Port:     appPort,
		Env:      getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	// JWT configuration
	jwtRefreshExpiration := getEnv("JWT_REFRESH_EXPIRATION_TIME", "168h")
	jwtAccessExpiration := getEnv("JWT_ACCESS_EXPIRATION_TIME", "1h")

	config.JWT = JWTConfig{
		Secret:            getEnv("JWT_SECRET_KEY", ""),
		RefreshExpiration: jwtRefreshExpiration,
		AccessExpiration:  jwtAccessExpiration,
	}

	// OAuth2 Google Configuration
	GoogleClientID := getEnv("CLIENT_ID", "")
	GoogleClientSecret := getEnv("CLIENT_SECRET", "")
	GoogleRedirectURL := getEnv("REDIRECT_URL", "")
	GoogleScopes := getEnvSlice("SCOPES")
	config.OAuth2Google = OAuth2GoogleConfig{
		ClientID:     GoogleClientID,
		ClientSecret: GoogleClientSecret,
		RedirectURL:  GoogleRedirectURL,
		Scopes:       GoogleScopes,
	}

	// Session configuration
	// sessionTimeout, err := time.ParseDuration(getEnv("SESSION_TIMEOUT", "30m"))
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid SESSION_TIMEOUT: %w", err)
	// }

	// config.Session = SessionConfig{
	// 	Secret:  getEnv("SESSION_SECRET", ""),
	// 	Timeout: sessionTimeout,
	// }

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.OAuth2Google.ClientID == "" {
		return fmt.Errorf("CLIENT_ID is required")
	}
	if c.OAuth2Google.ClientSecret == "" {
		return fmt.Errorf("CLIENT_SECRET is required")
	}

	if c.OAuth2Google.RedirectURL == "" {
		return fmt.Errorf("REDIRECT_URL is required")
	}

	if len(c.OAuth2Google.Scopes) == 0 {
		return fmt.Errorf("SCOPES is required")
	}
	// if c.Session.Secret == "" {
	// 	return fmt.Errorf("SESSION_SECRET is required")
	// }
	return nil
}

// DatabaseURL returns the PostgreSQL connection string
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvSlice(env string) []string {
	value := getEnv(env, "")
	if value == "" {
		return []string{}
	}
	var result []string = strings.Split(value, ",")
	return result
}
