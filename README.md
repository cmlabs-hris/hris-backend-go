# HRIS Backend Go

A full-featured **Human Resource Information System (HRIS)** backend built with Go, designed for multi-tenant company management with subscription-based feature gating, real-time notifications, and payment integration.

## Project Description

HRIS Backend Go provides a comprehensive REST API for managing every aspect of human resources — from employee onboarding and attendance tracking to payroll processing and leave management. The system is built around a **multi-tenant architecture** where each company operates in isolation with its own employees, schedules, and configurations.

### Why These Technologies?

- **Go** — Chosen for high concurrency, low memory footprint, and excellent performance for a backend serving multiple companies simultaneously.
- **Chi Router** — Lightweight, idiomatic Go HTTP router with composable middleware, ideal for the deeply nested route hierarchy and role-based access control.
- **PostgreSQL** — Robust relational database with strong transactional support, critical for payroll calculations and financial data integrity.
- **JWT + Google OAuth2** — Stateless authentication for scalability, with social login for frictionless employee onboarding.
- **Xendit** — Payment gateway integration for subscription billing and invoice management (popular in Southeast Asia markets).
- **Server-Sent Events (SSE)** — Lightweight real-time notification delivery without the overhead of WebSocket connections.

### Key Challenges & Design Decisions

- **Subscription-based feature gating** — Middleware dynamically enables/disables features (attendance, leave, payroll, scheduling) based on each company's active subscription plan and seat limits.
- **Multi-role authorization** — Four-tier access control (Owner → Manager → Employee → Pending) enforced at the middleware layer across all routes.
- **Automated background jobs** — Cron-based scheduling for subscription expiry checks and automated attendance record generation.

---

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation & Setup](#installation--setup)
- [Environment Variables](#environment-variables)
- [Running the Application](#running-the-application)
- [API Documentation](#api-documentation)
- [API Endpoints Overview](#api-endpoints-overview)
- [Example API Usage](#example-api-usage)
- [Running Tests](#running-tests)
- [Troubleshooting](#troubleshooting)
- [Credits](#credits)
- [License](#license)

---

## Features

### Core HR Modules
- **Authentication** — Email/password login, employee-code login, JWT access/refresh tokens, Google OAuth2, email verification, password reset
- **Company Management** — Multi-tenant company creation, profile management, logo upload
- **Employee Management** — Full CRUD, avatar upload, invitation-based onboarding, employee search and filtering
- **Attendance** — Clock-in/clock-out with geolocation, photo proof uploads, manager approval/rejection workflow
- **Leave Management** — Configurable leave types, quota allocation and adjustment, leave request approval workflow
- **Payroll** — Payroll components, employee component assignment, payroll generation, finalization, and summary reports
- **Work Schedules** — Flexible schedule definitions with time slots and location-based rules, employee schedule assignments

### Platform Features
- **Subscription & Billing** — Tiered plans with feature gating, Xendit payment integration, invoice management, seat-based pricing, plan upgrades/downgrades
- **Real-time Notifications** — Server-Sent Events (SSE) with notification preferences, batch processing, and read/unread tracking
- **Invitation System** — Token-based employee invitations via email with accept/reject workflow
- **Master Data** — Branches, grades, and positions management
- **Dashboards** — Admin dashboard (company-wide stats) and employee dashboard (personal work stats, attendance/leave summaries)
- **Reports** — Monthly attendance, payroll summary, leave balance, and new hire reports
- **Cron Jobs** — Automated subscription expiry checks and attendance record generation
- **File Storage** — Local file storage with MinIO/S3 migration path, supporting avatars, company logos, attendance photos, and leave attachments
- **Swagger UI** — Auto-served OpenAPI documentation at `/swagger/`

---

## Tech Stack

| Category | Technology |
|---|---|
| Language | Go 1.26+ |
| HTTP Router | [Chi v5](https://github.com/go-chi/chi) |
| Database | PostgreSQL 15+ |
| Database Driver | [pgx v5](https://github.com/jackc/pgx) |
| Authentication | JWT ([jwtauth](https://github.com/go-chi/jwtauth) + [jwx](https://github.com/lestrrat-go/jwx)) |
| OAuth2 | Google OAuth2 (`golang.org/x/oauth2`) |
| Payment Gateway | [Xendit Go SDK v7](https://github.com/xendit/xendit-go) |
| Email | SMTP (net/smtp) |
| API Docs | [Swaggo](https://github.com/swaggo/swag) + OpenAPI 3.0 |
| Logging | Structured JSON logging ([httplog](https://github.com/go-chi/httplog)) |
| Precision Math | [shopspring/decimal](https://github.com/shopspring/decimal) |
| Image Processing | `golang.org/x/image` |
| Config | [godotenv](https://github.com/joho/godotenv) |

---

## Architecture

The project follows a **Clean Architecture / Layered Architecture** pattern:

```
┌─────────────────────────────────────────────┐
│                  HTTP Layer                  │
│         (Handlers, Router, Middleware)       │
├─────────────────────────────────────────────┤
│                Service Layer                 │
│            (Business Logic)                  │
├─────────────────────────────────────────────┤
│                Domain Layer                  │
│       (Entities, DTOs, Interfaces)           │
├─────────────────────────────────────────────┤
│              Repository Layer                │
│        (PostgreSQL Implementations)          │
├─────────────────────────────────────────────┤
│            Infrastructure Layer              │
│     (Database, Storage, Email, Cron)         │
└─────────────────────────────────────────────┘
```

**Middleware Chain:** CORS → Logging → Content-Type → CleanPath → Recovery → Heartbeat → JWT Verification → Role Authorization → Subscription Feature Gate

---

## Project Structure

```
hris-backend-go/
├── cmd/
│   └── api/
│       └── main.go                  # Application entrypoint & dependency wiring
├── api/
│   ├── openapi.json                 # OpenAPI 3.0 specification
│   ├── postman_collection.json      # Postman collection for API testing
│   └── generate_postman.js          # Script to generate Postman collection from OpenAPI
├── internal/
│   ├── config/
│   │   └── config.go                # Environment variable loading & validation
│   ├── domain/                      # Domain entities, DTOs, and interfaces
│   │   ├── attendance/
│   │   ├── auth/
│   │   ├── company/
│   │   ├── dashboard/
│   │   ├── employee/
│   │   ├── employee_dashboard/
│   │   ├── invitation/
│   │   ├── leave/
│   │   ├── master/                  # Branches, grades, positions
│   │   ├── notification/
│   │   ├── payroll/
│   │   ├── report/
│   │   ├── schedule/
│   │   ├── subscription/
│   │   └── user/
│   ├── handler/
│   │   └── http/
│   │       ├── router.go            # Route definitions & middleware setup
│   │       ├── middleware/           # Auth, role, subscription middleware
│   │       ├── response/            # Standardized HTTP response helpers
│   │       ├── auth.go
│   │       ├── attendance.go
│   │       ├── company.go
│   │       ├── dashboard.go
│   │       ├── employee.go
│   │       ├── employee_dashboard.go
│   │       ├── invitation.go
│   │       ├── leave.go
│   │       ├── master.go
│   │       ├── notification.go
│   │       ├── payroll.go
│   │       ├── report.go
│   │       ├── schedule.go
│   │       └── subscription.go
│   ├── service/                     # Business logic layer
│   │   ├── attendance/
│   │   ├── auth/
│   │   ├── company/
│   │   ├── dashboard/
│   │   ├── employee/
│   │   ├── employee_dashboard/
│   │   ├── file/
│   │   ├── invitation/
│   │   ├── leave/
│   │   ├── master/
│   │   ├── notification/
│   │   ├── payroll/
│   │   ├── report/
│   │   ├── schedule/
│   │   └── subscription/
│   ├── repository/
│   │   └── postgresql/              # PostgreSQL query implementations
│   ├── infrastructure/
│   │   └── database/
│   │       └── postgresql/
│   │           └── migrations/      # SQL migration files (golang-migrate)
│   ├── fixtures/
│   │   └── company_defaults.go      # Default master data seeding
│   └── pkg/                         # Shared internal packages
│       ├── cron/                    # Background job scheduler
│       ├── database/                # Database connection pool
│       ├── email/                   # SMTP email service
│       ├── jwt/                     # JWT token service
│       ├── oauth/                   # Google OAuth2 service
│       ├── sse/                     # Server-Sent Events hub
│       ├── storage/                 # File storage abstraction (local / MinIO)
│       ├── utils/                   # Shared utilities
│       ├── validator/               # Request validation
│       └── xendit/                  # Xendit payment client & webhook verifier
├── storage/                         # Local file storage (avatars, logos, attendance, leave)
├── .env.example                     # Environment variable template
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## Prerequisites

- **Go 1.26+** (recommended: latest stable)
- **PostgreSQL 15+**
- **SMTP Server** (e.g., Gmail with app password) — for email invitations and password resets
- **[golang-migrate CLI](https://github.com/golang-migrate/migrate)** — for running database migrations
- **[Optional]** Google Cloud OAuth credentials — for Google social login
- **[Optional]** Xendit API key — for subscription billing (sandbox mode available)

---

## Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/cmlabs-hris/hris-backend-go.git
cd hris-backend-go
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Configure Environment Variables

Copy the example file and fill in the required values:

```bash
cp .env.example .env
```

See the [Environment Variables](#environment-variables) section for a full reference.

### 4. Create the Database

Connect to PostgreSQL and create the database:

```sql
CREATE DATABASE hris_db;
```

### 5. Run Database Migrations

Make sure PostgreSQL is running and the database exists, then run:

```bash
migrate -path ./internal/infrastructure/database/postgresql/migrations \
  -database "pgx5://hris_user:your_password@localhost:5432/hris_db?sslmode=disable" up
```

> **Note:** The project uses [golang-migrate](https://github.com/golang-migrate/migrate) with the `pgx5` driver. You can also apply the SQL files manually if preferred.

### 6. Run the Application

```bash
go run ./cmd/api
```

The server will start at `http://localhost:8080` (or the port configured in `APP_PORT`).

---

## Environment Variables

Copy `.env.example` to `.env` and configure the following:

| Variable | Description | Default |
|---|---|---|
| **Database** | | |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password (**required**) | — |
| `DB_NAME` | Database name | `cmlasb-hris` |
| `DB_SSL_MODE` | SSL mode | `disable` |
| **Application** | | |
| `APP_PORT` | Server port | `8080` |
| `APP_ENV` | Environment (`development` / `production`) | `development` |
| `LOG_LEVEL` | Log level | `info` |
| `FRONTEND_URL` | Frontend app URL (for CORS and redirects) | `http://localhost:3000` |
| **JWT** | | |
| `JWT_SECRET_KEY` | JWT signing secret (**required**) | — |
| `JWT_ACCESS_EXPIRATION_TIME` | Access token TTL | `1h` |
| `JWT_REFRESH_EXPIRATION_TIME` | Refresh token TTL | `24h` |
| **Google OAuth2** | | |
| `CLIENT_ID` | Google OAuth client ID (**required**) | — |
| `CLIENT_SECRET` | Google OAuth client secret (**required**) | — |
| `REDIRECT_URL` | OAuth callback URL (**required**) | — |
| `SCOPES` | OAuth scopes (comma-separated) | `email` |
| **Storage** | | |
| `STORAGE_TYPE` | Storage backend (`local`) | — |
| `BASE_PATH` | Local storage directory | `./storage` |
| `BASE_URL` | Public URL for file access | `http://localhost:8080/uploads` |
| **SMTP** | | |
| `SMTP_HOST` | SMTP server host | — |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USERNAME` | SMTP username | — |
| `SMTP_PASSWORD` | SMTP password | — |
| `SMTP_FROM` | Sender email address | `noreply@hris.com` |
| `SMTP_FROM_NAME` | Sender display name | `HRIS` |
| **Xendit (Payment)** | | |
| `XENDIT_API_KEY` | Xendit API key | — |
| `XENDIT_WEBHOOK_TOKEN` | Webhook signature verification token | — |
| `XENDIT_BASE_URL` | Xendit API base URL | `https://api.xendit.co` |
| `XENDIT_ENVIRONMENT` | `sandbox` or `production` | `sandbox` |
| `XENDIT_INVOICE_EXPIRY_HOURS` | Invoice expiry time | `24` |
| `XENDIT_SUCCESS_REDIRECT` | Post-payment success URL | `http://localhost:3000/subscription/success` |
| `XENDIT_FAILURE_REDIRECT` | Post-payment failure URL | `http://localhost:3000/subscription/failed` |
| **Invitation** | | |
| `INVITATION_BASE_URL` | Base URL for invitation links | `http://localhost:3000` |

---

## Running the Application

### Development

```bash
go run ./cmd/api
```

### Build & Run Binary

```bash
go build -o hris-server ./cmd/api
./hris-server
```

The server will output:

```
Server running at http://localhost:8080
```

---

## API Documentation

### Swagger UI

Visit **[http://localhost:8080/swagger/](http://localhost:8080/swagger/)** after starting the server.

### OpenAPI Specification

The raw OpenAPI 3.0 spec is available at:
- File: [`api/openapi.json`](api/openapi.json)
- Endpoint: `GET /openapi.json`

### Postman Collection

Import [`api/postman_collection.json`](api/postman_collection.json) into Postman for a ready-to-use API collection.

To regenerate the Postman collection from the OpenAPI spec:

```bash
node api/generate_postman.js
```

---

## API Endpoints Overview

All API endpoints are prefixed with `/api/v1`.

### Authentication (`/auth`)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/auth/register` | Register company admin | Public |
| `POST` | `/auth/login` | Login with email/password | Public |
| `POST` | `/auth/login/employee-code` | Login with employee code | Public |
| `GET` | `/auth/login/oauth/google` | Initiate Google OAuth login | Public |
| `GET` | `/auth/oauth/callback/google` | Google OAuth callback | Public |
| `POST` | `/auth/refresh` | Refresh access token | Public |
| `POST` | `/auth/logout` | Logout (invalidate tokens) | Public |
| `POST` | `/auth/forgot-password` | Request password reset email | Public |
| `POST` | `/auth/reset-password` | Reset password with token | Public |
| `POST` | `/auth/verify-email` | Verify email address | Public |

### Company (`/company`)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/company` | Create company (pending users only) | JWT + Pending |
| `GET` | `/company/my` | Get current company details | JWT + Subscription |
| `PUT` | `/company/my` | Update company | JWT + Owner |
| `DELETE` | `/company/my` | Delete company | JWT + Owner |
| `POST` | `/company/my/logo` | Upload company logo | JWT + Owner |

### Employees (`/employees`)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/employees` | List employees (with filters) | JWT + Manager |
| `GET` | `/employees/search` | Autocomplete search | JWT + Manager |
| `GET` | `/employees/{id}` | Get employee details | JWT |
| `POST` | `/employees` | Create employee (with invitation) | JWT + Manager + Feature |
| `PUT` | `/employees/{id}` | Update employee | JWT + Manager |
| `DELETE` | `/employees/{id}` | Soft delete employee | JWT + Manager |
| `POST` | `/employees/{id}/avatar` | Upload employee avatar | JWT |

### Attendance (`/attendance`)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/attendance/my` | Get my attendance records | JWT |
| `GET` | `/attendance/status` | Get current attendance status | JWT |
| `POST` | `/attendance/clock-in` | Clock in | JWT + Feature |
| `POST` | `/attendance/clock-out` | Clock out | JWT + Feature |
| `GET` | `/attendance` | List all (with filters) | JWT + Manager + Feature |
| `POST` | `/attendance/{id}/approve` | Approve attendance | JWT + Manager + Feature |
| `POST` | `/attendance/{id}/reject` | Reject attendance | JWT + Manager + Feature |

### Leave (`/leave`)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/leave/types` | List leave types | JWT |
| `POST` | `/leave/types` | Create leave type | JWT + Owner + Feature |
| `GET` | `/leave/quota/my` | Get my leave quota | JWT |
| `GET` | `/leave/quota` | List all quotas | JWT + Manager + Feature |
| `POST` | `/leave/quota/adjust` | Adjust employee quota | JWT + Manager + Feature |
| `GET` | `/leave/requests/my` | Get my leave requests | JWT |
| `POST` | `/leave/requests` | Create leave request | JWT + Feature |
| `POST` | `/leave/requests/{id}/approve` | Approve leave request | JWT + Manager + Feature |
| `POST` | `/leave/requests/{id}/reject` | Reject leave request | JWT + Manager + Feature |

### Schedule (`/schedule`)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/schedule` | List work schedules | JWT |
| `POST` | `/schedule` | Create work schedule | JWT + Owner + Feature |
| `GET` | `/employee-schedules` | List assignments | JWT |
| `POST` | `/employee-schedules` | Assign schedule | JWT + Manager + Feature |

### Payroll (`/payroll`)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/payroll/settings` | Get payroll settings | JWT + Manager |
| `PUT` | `/payroll/settings` | Update payroll settings | JWT + Owner + Feature |
| `GET` | `/payroll/components` | List payroll components | JWT + Manager |
| `POST` | `/payroll/generate` | Generate payroll | JWT + Manager + Feature |
| `POST` | `/payroll/finalize` | Finalize payroll period | JWT + Owner + Feature |

### Subscription (`/subscription`)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/plans` | List available plans | Public |
| `GET` | `/subscription/my` | Get current subscription | JWT |
| `POST` | `/subscription/checkout` | Checkout subscription | JWT + Owner |
| `POST` | `/subscription/upgrade` | Upgrade plan | JWT + Owner |
| `POST` | `/subscription/cancel` | Cancel subscription | JWT + Owner |
| `POST` | `/webhook/xendit` | Xendit payment webhook | Public (signature verified) |

### Other Endpoints

| Group | Key Endpoints | Auth |
|---|---|---|
| **Dashboard** | `GET /dashboard/admin`, `GET /dashboard/employee` | JWT + Manager / JWT |
| **Notifications** | `GET /notifications`, `GET /notifications/stream` (SSE) | JWT |
| **Reports** | `GET /reports/attendance`, `/reports/payroll`, `/reports/leave-balance`, `/reports/new-hires` | JWT + Manager |
| **Master Data** | CRUD for `/master/branches`, `/master/grades`, `/master/positions` | JWT (Manager for writes) |
| **Invitations** | `GET /invitations/my`, `POST /invitations/{token}/accept`, `GET /invitations/view/{token}` | JWT / Public |

---

## Example API Usage

### Register a Company Admin

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

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "email": "admin@acme.com",
    "password": "yourpassword"
}
```

### Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
    "refresh_token": "<your_refresh_token>"
}
```

### Clock In (requires `Authorization: Bearer <token>`)

```http
POST /api/v1/attendance/clock-in
Authorization: Bearer <access_token>
Content-Type: multipart/form-data

latitude: -6.2088
longitude: 106.8456
photo: <file>
```

---

## Running Tests

```bash
go test ./...
```

To run tests with verbose output:

```bash
go test -v ./...
```

To generate a coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Troubleshooting

| Problem | Solution |
|---|---|
| `Error loading .env file` | Ensure `.env` exists in the project root. Copy from `.env.example`. |
| `DB_PASSWORD is required` | Set `DB_PASSWORD` in your `.env` file. |
| `CLIENT_ID is required` | Set Google OAuth credentials in `.env`. For dev, you can create test credentials at [Google Cloud Console](https://console.cloud.google.com/apis/credentials). |
| Database connection error | Verify PostgreSQL is running, the database exists, and credentials are correct. |
| Migration errors | Ensure `golang-migrate` CLI is installed and the database URL uses the `pgx5://` scheme. |
| Port already in use | Change `APP_PORT` in `.env` or stop the conflicting process. |
| SMTP/email errors | Verify SMTP credentials. For Gmail, use an [App Password](https://support.google.com/accounts/answer/185833). |

---

## Credits

Built by the **cmlabs HRIS Team 5** — a project-based learning (PBL) initiative.

- Frontend: [cmlabs-hris-team5.vercel.app](https://cmlabs-hris-team5.vercel.app)
- Backend: This repository

---

## License

This project is licensed under the [MIT License](https://choosealicense.com/licenses/mit/).