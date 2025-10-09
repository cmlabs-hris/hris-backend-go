# hris-backend-go

HRIS (Human Resource Information System) backend service written in Go.

## Features

- Company and user management
- Employee management
- Authentication (JWT, refresh token, Google OAuth)
- Attendance, leave, and more (see database schema)

## Prerequisites

- Go 1.21+ (recommended: latest stable)
- PostgreSQL 15+
- [Optional] Redis (for caching/session, if enabled)
- [Optional] Google Cloud OAuth credentials (for Google login)

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/cmlabs-hris/hris-backend-go.git
cd hris-backend-go
```

### 2. Setup Environment Variables

Copy `.env.example` to `.env` and fill in the required values:

```bash
cp .env.example .env
```

Edit `.env` and set:

- Database credentials (`DB_HOST`, `DB_USER`, `DB_PASSWORD`, etc)
- JWT secret and expiration times
- Google OAuth credentials (if using Google login)

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run Database Migrations

Make sure PostgreSQL is running and accessible.

You can use any migration tool (e.g., [golang-migrate](https://github.com/golang-migrate/migrate)):

```bash
migrate -path ./internal/infrastructure/database/postgresql/migrations -database "pgx5://USER:PASSWORD@HOST:PORT/DBNAME?sslmode=disable" up
```

Or run the SQL files manually.

### 5. Run the Application

```bash
go run ./cmd/api
```

The server will start at `http://localhost:8080` (or as configured).

### 6. API Documentation

See [`api/openapi.json`](api/openapi.json) for OpenAPI documentation.

You can use Swagger UI or [Redoc](https://redocly.github.io/redoc/) to visualize the API.

### 7. Example API Usage

#### Register Company Admin

```http
POST /api/v1/auth/register
Content-Type: application/json

{
    "company_name": "Acme Corp",
    "company_username": "acmecorp",
    "email": "admin@acme.com",
    "password": "yourpassword",
    "confirm_password": "yourpassword"
}
```

#### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "email": "admin@acme.com",
    "password": "yourpassword"
}
```

#### Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
    "refresh_token": "<your_refresh_token>"
}
```

### 8. Project Structure

- `cmd/api/` - Application entrypoint
- `internal/` - Main application code (handlers, services, repositories, domain)
- `internal/infrastructure/database/postgresql/migrations/` - Database schema migrations
- `api/openapi.json` - OpenAPI spec

### 9. Running Tests

```bash
go test ./...
```

### 10. Troubleshooting

- Ensure your `.env` is configured correctly.
- Check database connectivity.
- For migration issues, verify your PostgreSQL version and permissions.

---

## License

MIT
