## Plan: Employee Invitation System (Final Complete)

**TL;DR:** Synchronous email dalam transaction, check duplicate email di pending invitation yang belum expired, minimal DB access dengan complex queries, semua filter by `company_id`.

---

### Architecture

```
┌─────────────────────┐
│   EmployeeService   │
│  - employeeRepo     │
│  - invitationSvc    │────▶┌─────────────────────┐
└─────────────────────┘     │  InvitationService  │
                            │  - invitationRepo   │
                            │  - employeeRepo     │
                            │  - companyRepo      │
                            │  - userRepo         │
                            │  - emailService     │
                            └─────────────────────┘
```

---

### Database Schema

```sql
CREATE TABLE employee_invitations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    invited_by_employee_id UUID NOT NULL REFERENCES employees(id),
    email VARCHAR(254) NOT NULL,
    token UUID UNIQUE NOT NULL DEFAULT uuidv7(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invitations_token ON employee_invitations(token);
CREATE INDEX idx_invitations_email_status ON employee_invitations(email, status);
CREATE INDEX idx_invitations_employee_id ON employee_invitations(employee_id);
CREATE INDEX idx_invitations_company_id ON employee_invitations(company_id);
```

---

### Steps

1. **Create invitation domain** at `internal/domain/invitation/`
   - `entity.go` - `Invitation`, `InvitationWithDetails` structs
   - `dto.go` - Request/Response DTOs
   - `errors.go` - Domain errors
   - `repository.go` - Repository interface
   - `service.go` - Service interface

2. **Add configs** at `internal/config/config.go`
   - `SMTPConfig` struct
   - `InvitationConfig` struct

3. **Create email package** at `internal/pkg/email/`
   - `email.go` - EmailService interface & SMTP implementation
   - `templates/invitation.html` - HTML email template

4. **Create invitation repository** at `internal/repository/postgresql/invitation.go`

5. **Create invitation service** at `internal/service/invitation/service.go`

6. **Create invitation handler** at `internal/handler/http/invitation.go`

7. **Update employee dto** at `internal/domain/employee/dto.go`
   - Add `Email` field to `CreateEmployeeRequest`

8. **Update employee repository** at `internal/repository/postgresql/employee.go`
   - Add `LinkUser` method

9. **Update employee service** at `internal/service/employee/service.go`
   - Inject `InvitationService`
   - Update `CreateEmployee` to auto-create invitation

10. **Update employee handler** at `internal/handler/http/employee.go`
    - Add `ResendInvitation`, `RevokeInvitation` methods

11. **Update router** at `internal/handler/http/router.go`

12. **Update main.go** at `cmd/api/main.go`

---

### Detailed Implementation

#### 1. Entity
**File:** `internal/domain/invitation/entity.go`

```go
package invitation

import "time"

type Status string

const (
    StatusPending  Status = "pending"
    StatusAccepted Status = "accepted"
    StatusRevoked  Status = "revoked"
)

type Invitation struct {
    ID                  string
    EmployeeID          string
    CompanyID           string
    InvitedByEmployeeID string
    Email               string
    Token               string
    Status              Status
    ExpiresAt           time.Time
    AcceptedAt          *time.Time
    RevokedAt           *time.Time
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

// For complex query results with JOINs
type InvitationWithDetails struct {
    Invitation
    EmployeeName  string
    CompanyName   string
    CompanyLogo   *string
    PositionName  *string
    InviterName   string
}

// Check if invitation is expired (query-time check)
func (i *Invitation) IsExpired() bool {
    return time.Now().After(i.ExpiresAt)
}

// Check if invitation can be accepted
func (i *Invitation) CanBeAccepted() bool {
    return i.Status == StatusPending && !i.IsExpired()
}
```

---

#### 2. DTO
**File:** `internal/domain/invitation/dto.go`

```go
package invitation

import (
    "github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

// CreateRequest - used internally by EmployeeService
type CreateRequest struct {
    EmployeeID          string
    CompanyID           string
    InvitedByEmployeeID string
    Email               string
    EmployeeName        string  // For email template
    InviterName         string  // For email template
    CompanyName         string  // For email template
    PositionName        *string // For email template
}

// AcceptRequest
type AcceptRequest struct {
    Token  string `json:"token" validate:"required,uuid"`
    UserID string // From JWT
}

func (r *AcceptRequest) Validate() error {
    return validator.Validate(r)
}

// MyInvitationResponse - GET /invitations/my
type MyInvitationResponse struct {
    Token        string  `json:"token"`
    CompanyName  string  `json:"company_name"`
    CompanyLogo  *string `json:"company_logo,omitempty"`
    PositionName *string `json:"position_name,omitempty"`
    InviterName  string  `json:"inviter_name"`
    ExpiresAt    string  `json:"expires_at"`
    CreatedAt    string  `json:"created_at"`
}

// InvitationDetailResponse - GET /invitations/{token}
type InvitationDetailResponse struct {
    Token        string  `json:"token"`
    Email        string  `json:"email"`
    EmployeeName string  `json:"employee_name"`
    CompanyName  string  `json:"company_name"`
    CompanyLogo  *string `json:"company_logo,omitempty"`
    PositionName *string `json:"position_name,omitempty"`
    InviterName  string  `json:"inviter_name"`
    Status       string  `json:"status"`
    ExpiresAt    string  `json:"expires_at"`
    IsExpired    bool    `json:"is_expired"`
}

// AcceptResponse
type AcceptResponse struct {
    Message     string `json:"message"`
    CompanyID   string `json:"company_id"`
    CompanyName string `json:"company_name"`
    EmployeeID  string `json:"employee_id"`
}
```

---

#### 3. Errors
**File:** `internal/domain/invitation/errors.go`

```go
package invitation

import "errors"

var (
    ErrInvitationNotFound      = errors.New("invitation not found")
    ErrInvitationExpired       = errors.New("invitation has expired")
    ErrInvitationAlreadyUsed   = errors.New("invitation has already been used")
    ErrInvitationRevoked       = errors.New("invitation has been revoked")
    ErrEmailMismatch           = errors.New("your email does not match the invitation email")
    ErrEmailAlreadyInvited     = errors.New("email already has a pending invitation in this company")
    ErrEmployeeAlreadyLinked   = errors.New("employee already linked to a user")
    ErrUserAlreadyHasCompany   = errors.New("user already belongs to a company")
    ErrCannotRevokeAccepted    = errors.New("cannot revoke an accepted invitation")
    ErrNoPendingInvitation     = errors.New("no pending invitation found for this employee")
)
```

---

#### 4. Repository Interface
**File:** `internal/domain/invitation/repository.go`

```go
package invitation

import (
    "context"
    "time"
)

type InvitationRepository interface {
    // Create new invitation
    Create(ctx context.Context, inv Invitation) (Invitation, error)
    
    // Get by token with all details (company name, inviter name, etc)
    GetByTokenWithDetails(ctx context.Context, token string) (InvitationWithDetails, error)
    
    // Get pending invitation by employee ID and company ID
    GetPendingByEmployeeID(ctx context.Context, employeeID, companyID string) (Invitation, error)
    
    // Check if email has pending non-expired invitation in company
    ExistsPendingByEmail(ctx context.Context, email, companyID string) (bool, error)
    
    // List pending non-expired invitations by email (for user's "my invitations")
    ListPendingByEmail(ctx context.Context, email string) ([]InvitationWithDetails, error)
    
    // Mark as accepted
    MarkAccepted(ctx context.Context, id string) error
    
    // Mark as revoked
    MarkRevoked(ctx context.Context, id string) error
    
    // Update token and expires_at (for resend)
    UpdateToken(ctx context.Context, id, newToken string, expiresAt time.Time) error
}
```

---

#### 5. Service Interface
**File:** `internal/domain/invitation/service.go`

```go
package invitation

import "context"

type InvitationService interface {
    // Create invitation and send email (called from EmployeeService)
    CreateAndSend(ctx context.Context, req CreateRequest) (Invitation, error)
    
    // Get invitation by token (public)
    GetByToken(ctx context.Context, token string) (InvitationDetailResponse, error)
    
    // List pending invitations for user's email
    ListMyInvitations(ctx context.Context, email string) ([]MyInvitationResponse, error)
    
    // Accept invitation (link user to employee)
    Accept(ctx context.Context, token, userID, userEmail string) (AcceptResponse, error)
    
    // Resend invitation email (generate new token)
    Resend(ctx context.Context, employeeID, companyID string) error
    
    // Revoke pending invitation
    Revoke(ctx context.Context, employeeID, companyID string) error
    
    // Check if email has pending invitation (for CreateEmployee validation)
    ExistsPendingByEmail(ctx context.Context, email, companyID string) (bool, error)
}
```

---

#### 6. Config
**File:** `internal/config/config.go` (additions)

```go
type SMTPConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    From     string
    FromName string
}

type InvitationConfig struct {
    ExpiryDays int
    BaseURL    string  // e.g., "https://app.hris.com"
}

type Config struct {
    // ... existing fields
    SMTP       SMTPConfig
    Invitation InvitationConfig
}

// In Load() function:
smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
expiryDays, _ := strconv.Atoi(getEnv("INVITATION_EXPIRY_DAYS", "7"))

config.SMTP = SMTPConfig{
    Host:     getEnv("SMTP_HOST", ""),
    Port:     smtpPort,
    Username: getEnv("SMTP_USERNAME", ""),
    Password: getEnv("SMTP_PASSWORD", ""),
    From:     getEnv("SMTP_FROM", "noreply@hris.com"),
    FromName: getEnv("SMTP_FROM_NAME", "HRIS"),
}

config.Invitation = InvitationConfig{
    ExpiryDays: expiryDays,
    BaseURL:    getEnv("INVITATION_BASE_URL", "http://localhost:3000"),
}
```

---

#### 7. Email Service
**File:** `internal/pkg/email/email.go`

```go
package email

import (
    "bytes"
    "embed"
    "fmt"
    "html/template"
    "net/smtp"
    
    "github.com/cmlabs-hris/hris-backend-go/internal/config"
)

//go:embed templates/*.html
var templateFS embed.FS

type EmailService interface {
    SendInvitation(to, employeeName, inviterName, companyName string, positionName *string, invitationLink, expiresAt string) error
}

type emailServiceImpl struct {
    cfg       config.SMTPConfig
    templates *template.Template
}

func NewEmailService(cfg config.SMTPConfig) (EmailService, error) {
    tmpl, err := template.ParseFS(templateFS, "templates/*.html")
    if err != nil {
        return nil, fmt.Errorf("failed to parse email templates: %w", err)
    }
    
    return &emailServiceImpl{
        cfg:       cfg,
        templates: tmpl,
    }, nil
}

type invitationEmailData struct {
    EmployeeName   string
    InviterName    string
    CompanyName    string
    PositionName   string
    InvitationLink string
    ExpiresAt      string
}

func (s *emailServiceImpl) SendInvitation(to, employeeName, inviterName, companyName string, positionName *string, invitationLink, expiresAt string) error {
    data := invitationEmailData{
        EmployeeName:   employeeName,
        InviterName:    inviterName,
        CompanyName:    companyName,
        PositionName:   "",
        InvitationLink: invitationLink,
        ExpiresAt:      expiresAt,
    }
    if positionName != nil {
        data.PositionName = *positionName
    }
    
    var body bytes.Buffer
    if err := s.templates.ExecuteTemplate(&body, "invitation.html", data); err != nil {
        return fmt.Errorf("failed to execute template: %w", err)
    }
    
    return s.sendHTML(to, fmt.Sprintf("Undangan Bergabung ke %s", companyName), body.String())
}

func (s *emailServiceImpl) sendHTML(to, subject, htmlBody string) error {
    from := s.cfg.From
    
    headers := fmt.Sprintf("From: %s <%s>\r\n", s.cfg.FromName, from)
    headers += fmt.Sprintf("To: %s\r\n", to)
    headers += fmt.Sprintf("Subject: %s\r\n", subject)
    headers += "MIME-Version: 1.0\r\n"
    headers += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
    headers += "\r\n"
    
    message := []byte(headers + htmlBody)
    
    auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
    addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
    
    return smtp.SendMail(addr, auth, from, []string{to}, message)
}
```

---

#### 8. HTML Template
**File:** `internal/pkg/email/templates/invitation.html`

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Undangan Bergabung</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .content { padding: 30px; }
        .greeting { font-size: 18px; color: #333; margin-bottom: 20px; }
        .message { color: #666; line-height: 1.6; margin-bottom: 25px; }
        .highlight { background: #f8f9fa; border-left: 4px solid #667eea; padding: 15px; margin: 20px 0; }
        .highlight p { margin: 5px 0; color: #333; }
        .button-container { text-align: center; margin: 30px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; text-decoration: none; padding: 15px 40px; border-radius: 5px; font-weight: bold; font-size: 16px; }
        .button:hover { opacity: 0.9; }
        .expiry { color: #999; font-size: 14px; text-align: center; margin-top: 20px; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        .warning { color: #999; font-size: 13px; margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Undangan Bergabung</h1>
        </div>
        <div class="content">
            <p class="greeting">Halo <strong>{{.EmployeeName}}</strong>!</p>
            
            <p class="message">
                <strong>{{.InviterName}}</strong> dari <strong>{{.CompanyName}}</strong> mengundang Anda untuk bergabung sebagai karyawan.
            </p>
            
            {{if .PositionName}}
            <div class="highlight">
                <p><strong>Posisi:</strong> {{.PositionName}}</p>
                <p><strong>Perusahaan:</strong> {{.CompanyName}}</p>
            </div>
            {{end}}
            
            <div class="button-container">
                <a href="{{.InvitationLink}}" class="button">Terima Undangan</a>
            </div>
            
            <p class="expiry">Link ini akan kadaluarsa pada <strong>{{.ExpiresAt}}</strong></p>
            
            <p class="warning">
                Jika Anda tidak mengenal pengirim atau tidak mengharapkan undangan ini, 
                Anda dapat mengabaikan email ini dengan aman.
            </p>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis oleh sistem HRIS.</p>
            <p>Mohon jangan membalas email ini.</p>
        </div>
    </div>
</body>
</html>
```

---

#### 9. Invitation Repository
**File:** `internal/repository/postgresql/invitation.go`

```go
package postgresql

import (
    "context"
    "time"
    
    "github.com/cmlabs-hris/hris-backend-go/internal/domain/invitation"
    "github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
    "github.com/jackc/pgx/v5"
)

type invitationRepositoryImpl struct {
    db *database.DB
}

func NewInvitationRepository(db *database.DB) invitation.InvitationRepository {
    return &invitationRepositoryImpl{db: db}
}

func (r *invitationRepositoryImpl) Create(ctx context.Context, inv invitation.Invitation) (invitation.Invitation, error) {
    q := GetQuerier(ctx, r.db)
    
    query := `
        INSERT INTO employee_invitations (
            employee_id, company_id, invited_by_employee_id, email, token, status, expires_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, employee_id, company_id, invited_by_employee_id, email, token, status, 
                  expires_at, accepted_at, revoked_at, created_at, updated_at
    `
    
    var created invitation.Invitation
    err := q.QueryRow(ctx, query,
        inv.EmployeeID, inv.CompanyID, inv.InvitedByEmployeeID,
        inv.Email, inv.Token, inv.Status, inv.ExpiresAt,
    ).Scan(
        &created.ID, &created.EmployeeID, &created.CompanyID, &created.InvitedByEmployeeID,
        &created.Email, &created.Token, &created.Status, &created.ExpiresAt,
        &created.AcceptedAt, &created.RevokedAt, &created.CreatedAt, &created.UpdatedAt,
    )
    
    return created, err
}

func (r *invitationRepositoryImpl) GetByTokenWithDetails(ctx context.Context, token string) (invitation.InvitationWithDetails, error) {
    q := GetQuerier(ctx, r.db)
    
    query := `
        SELECT 
            ei.id, ei.employee_id, ei.company_id, ei.invited_by_employee_id, 
            ei.email, ei.token, ei.status, ei.expires_at, 
            ei.accepted_at, ei.revoked_at, ei.created_at, ei.updated_at,
            e.full_name as employee_name,
            c.name as company_name, c.logo_url as company_logo,
            p.name as position_name,
            inviter.full_name as inviter_name
        FROM employee_invitations ei
        JOIN employees e ON e.id = ei.employee_id
        JOIN companies c ON c.id = ei.company_id
        JOIN employees inviter ON inviter.id = ei.invited_by_employee_id
        LEFT JOIN positions p ON p.id = e.position_id
        WHERE ei.token = $1
    `
    
    var inv invitation.InvitationWithDetails
    var positionName *string
    
    err := q.QueryRow(ctx, query, token).Scan(
        &inv.ID, &inv.EmployeeID, &inv.CompanyID, &inv.InvitedByEmployeeID,
        &inv.Email, &inv.Token, &inv.Status, &inv.ExpiresAt,
        &inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt, &inv.UpdatedAt,
        &inv.EmployeeName, &inv.CompanyName, &inv.CompanyLogo,
        &positionName, &inv.InviterName,
    )
    if err != nil {
        if err == pgx.ErrNoRows {
            return inv, invitation.ErrInvitationNotFound
        }
        return inv, err
    }
    
    inv.PositionName = positionName
    return inv, nil
}

func (r *invitationRepositoryImpl) GetPendingByEmployeeID(ctx context.Context, employeeID, companyID string) (invitation.Invitation, error) {
    q := GetQuerier(ctx, r.db)
    
    query := `
        SELECT id, employee_id, company_id, invited_by_employee_id, email, token, status,
               expires_at, accepted_at, revoked_at, created_at, updated_at
        FROM employee_invitations
        WHERE employee_id = $1 AND company_id = $2 AND status = 'pending'
        ORDER BY created_at DESC
        LIMIT 1
    `
    
    var inv invitation.Invitation
    err := q.QueryRow(ctx, query, employeeID, companyID).Scan(
        &inv.ID, &inv.EmployeeID, &inv.CompanyID, &inv.InvitedByEmployeeID,
        &inv.Email, &inv.Token, &inv.Status, &inv.ExpiresAt,
        &inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt, &inv.UpdatedAt,
    )
    if err != nil {
        if err == pgx.ErrNoRows {
            return inv, invitation.ErrNoPendingInvitation
        }
        return inv, err
    }
    
    return inv, nil
}

func (r *invitationRepositoryImpl) ExistsPendingByEmail(ctx context.Context, email, companyID string) (bool, error) {
    q := GetQuerier(ctx, r.db)
    
    query := `
        SELECT EXISTS(
            SELECT 1 FROM employee_invitations 
            WHERE email = $1 AND company_id = $2 AND status = 'pending' AND expires_at > NOW()
        )
    `
    
    var exists bool
    err := q.QueryRow(ctx, query, email, companyID).Scan(&exists)
    return exists, err
}

func (r *invitationRepositoryImpl) ListPendingByEmail(ctx context.Context, email string) ([]invitation.InvitationWithDetails, error) {
    q := GetQuerier(ctx, r.db)
    
    query := `
        SELECT 
            ei.id, ei.employee_id, ei.company_id, ei.invited_by_employee_id, 
            ei.email, ei.token, ei.status, ei.expires_at, 
            ei.accepted_at, ei.revoked_at, ei.created_at, ei.updated_at,
            e.full_name as employee_name,
            c.name as company_name, c.logo_url as company_logo,
            p.name as position_name,
            inviter.full_name as inviter_name
        FROM employee_invitations ei
        JOIN employees e ON e.id = ei.employee_id
        JOIN companies c ON c.id = ei.company_id
        JOIN employees inviter ON inviter.id = ei.invited_by_employee_id
        LEFT JOIN positions p ON p.id = e.position_id
        WHERE ei.email = $1 AND ei.status = 'pending' AND ei.expires_at > NOW()
        ORDER BY ei.created_at DESC
    `
    
    rows, err := q.Query(ctx, query, email)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var invitations []invitation.InvitationWithDetails
    for rows.Next() {
        var inv invitation.InvitationWithDetails
        var positionName *string
        
        err := rows.Scan(
            &inv.ID, &inv.EmployeeID, &inv.CompanyID, &inv.InvitedByEmployeeID,
            &inv.Email, &inv.Token, &inv.Status, &inv.ExpiresAt,
            &inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt, &inv.UpdatedAt,
            &inv.EmployeeName, &inv.CompanyName, &inv.CompanyLogo,
            &positionName, &inv.InviterName,
        )
        if err != nil {
            return nil, err
        }
        
        inv.PositionName = positionName
        invitations = append(invitations, inv)
    }
    
    return invitations, rows.Err()
}

func (r *invitationRepositoryImpl) MarkAccepted(ctx context.Context, id string) error {
    q := GetQuerier(ctx, r.db)
    
    query := `
        UPDATE employee_invitations 
        SET status = 'accepted', accepted_at = NOW(), updated_at = NOW()
        WHERE id = $1
    `
    
    _, err := q.Exec(ctx, query, id)
    return err
}

func (r *invitationRepositoryImpl) MarkRevoked(ctx context.Context, id string) error {
    q := GetQuerier(ctx, r.db)
    
    query := `
        UPDATE employee_invitations 
        SET status = 'revoked', revoked_at = NOW(), updated_at = NOW()
        WHERE id = $1
    `
    
    _, err := q.Exec(ctx, query, id)
    return err
}

func (r *invitationRepositoryImpl) UpdateToken(ctx context.Context, id, newToken string, expiresAt time.Time) error {
    q := GetQuerier(ctx, r.db)
    
    query := `
        UPDATE employee_invitations 
        SET token = $1, expires_at = $2, updated_at = NOW()
        WHERE id = $3
    `
    
    _, err := q.Exec(ctx, query, newToken, expiresAt, id)
    return err
}
```

---

#### 10. Invitation Service
**File:** `internal/service/invitation/service.go`

```go
package invitation

import (
    "context"
    "fmt"
    "time"
    
    "github.com/cmlabs-hris/hris-backend-go/internal/config"
    "github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
    "github.com/cmlabs-hris/hris-backend-go/internal/domain/invitation"
    "github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
    "github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
    "github.com/cmlabs-hris/hris-backend-go/internal/pkg/email"
    "github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
)

type InvitationServiceImpl struct {
    db             *database.DB
    invitationRepo invitation.InvitationRepository
    employeeRepo   employee.EmployeeRepository
    userRepo       user.UserRepository
    emailService   email.EmailService
    config         config.InvitationConfig
}

func NewInvitationService(
    db *database.DB,
    invitationRepo invitation.InvitationRepository,
    employeeRepo employee.EmployeeRepository,
    userRepo user.UserRepository,
    emailService email.EmailService,
    config config.InvitationConfig,
) invitation.InvitationService {
    return &InvitationServiceImpl{
        db:             db,
        invitationRepo: invitationRepo,
        employeeRepo:   employeeRepo,
        userRepo:       userRepo,
        emailService:   emailService,
        config:         config,
    }
}

func (s *InvitationServiceImpl) CreateAndSend(ctx context.Context, req invitation.CreateRequest) (invitation.Invitation, error) {
    // Generate UUIDv7 token
    token := uuid.Must(uuid.NewV7()).String()
    expiresAt := time.Now().AddDate(0, 0, s.config.ExpiryDays)
    
    inv := invitation.Invitation{
        EmployeeID:          req.EmployeeID,
        CompanyID:           req.CompanyID,
        InvitedByEmployeeID: req.InvitedByEmployeeID,
        Email:               req.Email,
        Token:               token,
        Status:              invitation.StatusPending,
        ExpiresAt:           expiresAt,
    }
    
    // Create invitation record
    created, err := s.invitationRepo.Create(ctx, inv)
    if err != nil {
        return invitation.Invitation{}, fmt.Errorf("failed to create invitation: %w", err)
    }
    
    // Build invitation link
    invitationLink := fmt.Sprintf("%s/invitations/%s", s.config.BaseURL, token)
    expiresAtStr := expiresAt.Format("02 January 2006, 15:04 WIB")
    
    // Send email (synchronous)
    err = s.emailService.SendInvitation(
        req.Email,
        req.EmployeeName,
        req.InviterName,
        req.CompanyName,
        req.PositionName,
        invitationLink,
        expiresAtStr,
    )
    if err != nil {
        return invitation.Invitation{}, fmt.Errorf("failed to send invitation email: %w", err)
    }
    
    return created, nil
}

func (s *InvitationServiceImpl) GetByToken(ctx context.Context, token string) (invitation.InvitationDetailResponse, error) {
    inv, err := s.invitationRepo.GetByTokenWithDetails(ctx, token)
    if err != nil {
        return invitation.InvitationDetailResponse{}, err
    }
    
    return invitation.InvitationDetailResponse{
        Token:        inv.Token,
        Email:        inv.Email,
        EmployeeName: inv.EmployeeName,
        CompanyName:  inv.CompanyName,
        CompanyLogo:  inv.CompanyLogo,
        PositionName: inv.PositionName,
        InviterName:  inv.InviterName,
        Status:       string(inv.Status),
        ExpiresAt:    inv.ExpiresAt.Format("2006-01-02 15:04:05"),
        IsExpired:    inv.IsExpired(),
    }, nil
}

func (s *InvitationServiceImpl) ListMyInvitations(ctx context.Context, email string) ([]invitation.MyInvitationResponse, error) {
    invitations, err := s.invitationRepo.ListPendingByEmail(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("failed to list invitations: %w", err)
    }
    
    results := make([]invitation.MyInvitationResponse, 0, len(invitations))
    for _, inv := range invitations {
        results = append(results, invitation.MyInvitationResponse{
            Token:        inv.Token,
            CompanyName:  inv.CompanyName,
            CompanyLogo:  inv.CompanyLogo,
            PositionName: inv.PositionName,
            InviterName:  inv.InviterName,
            ExpiresAt:    inv.ExpiresAt.Format("2006-01-02 15:04:05"),
            CreatedAt:    inv.CreatedAt.Format("2006-01-02 15:04:05"),
        })
    }
    
    return results, nil
}

func (s *InvitationServiceImpl) Accept(ctx context.Context, token, userID, userEmail string) (invitation.AcceptResponse, error) {
    // Get invitation with details
    inv, err := s.invitationRepo.GetByTokenWithDetails(ctx, token)
    if err != nil {
        return invitation.AcceptResponse{}, err
    }
    
    // Validate invitation status
    if inv.Status == invitation.StatusAccepted {
        return invitation.AcceptResponse{}, invitation.ErrInvitationAlreadyUsed
    }
    if inv.Status == invitation.StatusRevoked {
        return invitation.AcceptResponse{}, invitation.ErrInvitationRevoked
    }
    if inv.IsExpired() {
        return invitation.AcceptResponse{}, invitation.ErrInvitationExpired
    }
    
    // Validate email matches
    if userEmail != inv.Email {
        return invitation.AcceptResponse{}, invitation.ErrEmailMismatch
    }
    
    // Get user to check if already has company
    userData, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return invitation.AcceptResponse{}, fmt.Errorf("failed to get user: %w", err)
    }
    if userData.CompanyID != nil && *userData.CompanyID != "" {
        return invitation.AcceptResponse{}, invitation.ErrUserAlreadyHasCompany
    }
    
    // Get employee to check if already linked
    emp, err := s.employeeRepo.GetByID(ctx, inv.EmployeeID)
    if err != nil {
        return invitation.AcceptResponse{}, fmt.Errorf("failed to get employee: %w", err)
    }
    if emp.UserID != "" {
        return invitation.AcceptResponse{}, invitation.ErrEmployeeAlreadyLinked
    }
    
    // Transaction: link user to employee, update user company/role, mark invitation accepted
    err = postgresql.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
        txCtx := context.WithValue(ctx, "tx", tx)
        
        // 1. Link employee to user
        if err := s.employeeRepo.LinkUser(txCtx, inv.EmployeeID, userID); err != nil {
            return fmt.Errorf("failed to link user to employee: %w", err)
        }
        
        // 2. Update user's company_id and role
        if err := s.userRepo.UpdateCompanyAndRole(txCtx, userID, inv.CompanyID, user.RoleEmployee); err != nil {
            return fmt.Errorf("failed to update user: %w", err)
        }
        
        // 3. Mark invitation as accepted
        if err := s.invitationRepo.MarkAccepted(txCtx, inv.ID); err != nil {
            return fmt.Errorf("failed to mark invitation accepted: %w", err)
        }
        
        return nil
    })
    if err != nil {
        return invitation.AcceptResponse{}, err
    }
    
    return invitation.AcceptResponse{
        Message:     "Invitation accepted successfully",
        CompanyID:   inv.CompanyID,
        CompanyName: inv.CompanyName,
        EmployeeID:  inv.EmployeeID,
    }, nil
}

func (s *InvitationServiceImpl) Resend(ctx context.Context, employeeID, companyID string) error {
    // Get pending invitation
    inv, err := s.invitationRepo.GetPendingByEmployeeID(ctx, employeeID, companyID)
    if err != nil {
        return err
    }
    
    // Generate new token and expiry
    newToken := uuid.Must(uuid.NewV7()).String()
    newExpiresAt := time.Now().AddDate(0, 0, s.config.ExpiryDays)
    
    // Update token
    if err := s.invitationRepo.UpdateToken(ctx, inv.ID, newToken, newExpiresAt); err != nil {
        return fmt.Errorf("failed to update token: %w", err)
    }
    
    // Get details for email
    invWithDetails, err := s.invitationRepo.GetByTokenWithDetails(ctx, newToken)
    if err != nil {
        return fmt.Errorf("failed to get invitation details: %w", err)
    }
    
    // Build invitation link
    invitationLink := fmt.Sprintf("%s/invitations/%s", s.config.BaseURL, newToken)
    expiresAtStr := newExpiresAt.Format("02 January 2006, 15:04 WIB")
    
    // Send email
    return s.emailService.SendInvitation(
        invWithDetails.Email,
        invWithDetails.EmployeeName,
        invWithDetails.InviterName,
        invWithDetails.CompanyName,
        invWithDetails.PositionName,
        invitationLink,
        expiresAtStr,
    )
}

func (s *InvitationServiceImpl) Revoke(ctx context.Context, employeeID, companyID string) error {
    // Get pending invitation
    inv, err := s.invitationRepo.GetPendingByEmployeeID(ctx, employeeID, companyID)
    if err != nil {
        return err
    }
    
    // Check if already accepted
    if inv.Status == invitation.StatusAccepted {
        return invitation.ErrCannotRevokeAccepted
    }
    
    return s.invitationRepo.MarkRevoked(ctx, inv.ID)
}

func (s *InvitationServiceImpl) ExistsPendingByEmail(ctx context.Context, email, companyID string) (bool, error) {
    return s.invitationRepo.ExistsPendingByEmail(ctx, email, companyID)
}
```

---

#### 11. Invitation Handler
**File:** `internal/handler/http/invitation.go`

```go
package http

import (
    "net/http"
    
    "github.com/cmlabs-hris/hris-backend-go/internal/domain/invitation"
    "github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/jwtauth/v5"
)

type InvitationHandler interface {
    GetByToken(w http.ResponseWriter, r *http.Request)
    ListMyInvitations(w http.ResponseWriter, r *http.Request)
    Accept(w http.ResponseWriter, r *http.Request)
}

type invitationHandlerImpl struct {
    invitationService invitation.InvitationService
}

func NewInvitationHandler(invitationService invitation.InvitationService) InvitationHandler {
    return &invitationHandlerImpl{
        invitationService: invitationService,
    }
}

// GetByToken - Public endpoint
func (h *invitationHandlerImpl) GetByToken(w http.ResponseWriter, r *http.Request) {
    token := chi.URLParam(r, "token")
    if token == "" {
        response.BadRequest(w, "Token is required", nil)
        return
    }
    
    result, err := h.invitationService.GetByToken(r.Context(), token)
    if err != nil {
        response.HandleError(w, err)
        return
    }
    
    response.Success(w, result)
}

// ListMyInvitations - Authenticated endpoint
func (h *invitationHandlerImpl) ListMyInvitations(w http.ResponseWriter, r *http.Request) {
    _, claims, err := jwtauth.FromContext(r.Context())
    if err != nil {
        response.Unauthorized(w, "Invalid token")
        return
    }
    
    email, ok := claims["email"].(string)
    if !ok || email == "" {
        response.BadRequest(w, "Email not found in token", nil)
        return
    }
    
    results, err := h.invitationService.ListMyInvitations(r.Context(), email)
    if err != nil {
        response.HandleError(w, err)
        return
    }
    
    response.Success(w, results)
}

// Accept - Authenticated endpoint
func (h *invitationHandlerImpl) Accept(w http.ResponseWriter, r *http.Request) {
    token := chi.URLParam(r, "token")
    if token == "" {
        response.BadRequest(w, "Token is required", nil)
        return
    }
    
    _, claims, err := jwtauth.FromContext(r.Context())
    if err != nil {
        response.Unauthorized(w, "Invalid token")
        return
    }
    
    userID, ok := claims["user_id"].(string)
    if !ok || userID == "" {
        response.BadRequest(w, "User ID not found in token", nil)
        return
    }
    
    email, ok := claims["email"].(string)
    if !ok || email == "" {
        response.BadRequest(w, "Email not found in token", nil)
        return
    }
    
    result, err := h.invitationService.Accept(r.Context(), token, userID, email)
    if err != nil {
        response.HandleError(w, err)
        return
    }
    
    response.SuccessWithMessage(w, result.Message, result)
}
```

---

#### 12. Employee DTO Update
**File:** `internal/domain/employee/dto.go` (modification)

```go
// Add Email field to CreateEmployeeRequest
type CreateEmployeeRequest struct {
    Email        string  `json:"email" validate:"required,email"`  // NEW - Required for invitation
    EmployeeCode string  `json:"employee_code" validate:"required,min=1,max=50"`
    FullName     string  `json:"full_name" validate:"required,min=1,max=255"`
    // ... rest unchanged
}
```

---

#### 13. Employee Repository Update
**File:** `internal/repository/postgresql/employee.go` (addition)

```go
// Add LinkUser method
func (e *employeeRepositoryImpl) LinkUser(ctx context.Context, employeeID, userID string) error {
    q := GetQuerier(ctx, e.db)
    
    query := `
        UPDATE employees 
        SET user_id = $1, updated_at = NOW()
        WHERE id = $2 AND user_id IS NULL
        RETURNING id
    `
    
    var id string
    err := q.QueryRow(ctx, query, userID, employeeID).Scan(&id)
    if err != nil {
        if err == pgx.ErrNoRows {
            return invitation.ErrEmployeeAlreadyLinked
        }
        return err
    }
    
    return nil
}
```

**File:** `internal/domain/employee/repository.go` (addition)

```go
// Add to interface
type EmployeeRepository interface {
    // ... existing methods
    LinkUser(ctx context.Context, employeeID, userID string) error
}
```

---

#### 14. Employee Service Update
**File:** `internal/service/employee/service.go` (modification)

```go
type EmployeeServiceImpl struct {
    db                *database.DB
    employeeRepo      employee.EmployeeRepository
    invitationService invitation.InvitationService  // NEW
    companyRepo       company.CompanyRepository     // NEW - for company name
    fileService       file.FileService
}

func NewEmployeeService(
    db *database.DB,
    employeeRepo employee.EmployeeRepository,
    invitationService invitation.InvitationService,
    companyRepo company.CompanyRepository,
    fileService file.FileService,
) employee.EmployeeService {
    return &EmployeeServiceImpl{
        db:                db,
        employeeRepo:      employeeRepo,
        invitationService: invitationService,
        companyRepo:       companyRepo,
        fileService:       fileService,
    }
}

func (s *EmployeeServiceImpl) CreateEmployee(ctx context.Context, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
    if err := req.Validate(); err != nil {
        return employee.EmployeeResponse{}, err
    }

    companyID, inviterEmployeeID, _, err := getClaimsFromContext(ctx)
    if err != nil {
        return employee.EmployeeResponse{}, err
    }

    // Check if email already has pending invitation (1 query)
    exists, err := s.invitationService.ExistsPendingByEmail(ctx, req.Email, companyID)
    if err != nil {
        return employee.EmployeeResponse{}, fmt.Errorf("failed to check email: %w", err)
    }
    if exists {
        return employee.EmployeeResponse{}, invitation.ErrEmailAlreadyInvited
    }

    // Check employee code and NIK (1 query with OR)
    exists, err = s.employeeRepo.ExistsByIDOrCodeOrNIK(ctx, companyID, nil, &req.EmployeeCode, req.NIK)
    if err != nil {
        return employee.EmployeeResponse{}, fmt.Errorf("failed to check duplicates: %w", err)
    }
    if exists {
        return employee.EmployeeResponse{}, employee.ErrEmployeeCodeExists
    }

    // Get inviter and company info for email (1 query with JOIN)
    inviter, err := s.employeeRepo.GetByID(ctx, inviterEmployeeID)
    if err != nil {
        return employee.EmployeeResponse{}, fmt.Errorf("failed to get inviter: %w", err)
    }
    
    companyData, err := s.companyRepo.GetByID(ctx, companyID)
    if err != nil {
        return employee.EmployeeResponse{}, fmt.Errorf("failed to get company: %w", err)
    }

    // Parse dates, prepare data...
    // (same as before)

    newEmployee := employee.Employee{
        // UserID is empty - will be linked when invitation accepted
        CompanyID:    companyID,
        // ... rest same
    }

    var response employee.EmployeeResponse
    
    // Transaction: create employee + create invitation + send email
    err = postgresql.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
        txCtx := context.WithValue(ctx, "tx", tx)

        // 1. Create employee (1 query)
        created, err := s.employeeRepo.Create(txCtx, newEmployee)
        if err != nil {
            return fmt.Errorf("failed to create employee: %w", err)
        }

        // 2. Get full details with JOINs (1 query)
        empWithDetails, err := s.employeeRepo.GetByIDWithDetails(txCtx, created.ID, companyID)
        if err != nil {
            return fmt.Errorf("failed to get employee details: %w", err)
        }
        response = mapEmployeeToResponse(empWithDetails)

        // 3. Create invitation and send email (1 query + SMTP)
        _, err = s.invitationService.CreateAndSend(txCtx, invitation.CreateRequest{
            EmployeeID:          created.ID,
            CompanyID:           companyID,
            InvitedByEmployeeID: inviterEmployeeID,
            Email:               req.Email,
            EmployeeName:        req.FullName,
            InviterName:         inviter.FullName,
            CompanyName:         companyData.Name,
            PositionName:        empWithDetails.PositionName,
        })
        if err != nil {
            return err  // Will rollback everything
        }

        return nil
    })

    if err != nil {
        return employee.EmployeeResponse{}, err
    }

    return response, nil
}
```

---

#### 15. Employee Handler Update
**File:** `internal/handler/http/employee.go` (additions)

```go
type EmployeeHandler interface {
    // ... existing methods
    ResendInvitation(w http.ResponseWriter, r *http.Request)
    RevokeInvitation(w http.ResponseWriter, r *http.Request)
}

type employeeHandlerImpl struct {
    employeeService   employee.EmployeeService
    invitationService invitation.InvitationService  // NEW
}

func NewEmployeeHandler(
    employeeService employee.EmployeeService,
    invitationService invitation.InvitationService,
) EmployeeHandler {
    return &employeeHandlerImpl{
        employeeService:   employeeService,
        invitationService: invitationService,
    }
}

// ResendInvitation - Manager+ only
func (h *employeeHandlerImpl) ResendInvitation(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        response.BadRequest(w, "Employee ID is required", nil)
        return
    }
    
    _, claims, _ := jwtauth.FromContext(r.Context())
    companyID, _ := claims["company_id"].(string)
    
    err := h.invitationService.Resend(r.Context(), id, companyID)
    if err != nil {
        response.HandleError(w, err)
        return
    }
    
    response.SuccessWithMessage(w, "Invitation resent successfully", nil)
}

// RevokeInvitation - Manager+ only
func (h *employeeHandlerImpl) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        response.BadRequest(w, "Employee ID is required", nil)
        return
    }
    
    _, claims, _ := jwtauth.FromContext(r.Context())
    companyID, _ := claims["company_id"].(string)
    
    err := h.invitationService.Revoke(r.Context(), id, companyID)
    if err != nil {
        response.HandleError(w, err)
        return
    }
    
    response.SuccessWithMessage(w, "Invitation revoked successfully", nil)
}
```

---

#### 16. Router Update
**File:** `internal/handler/http/router.go`

```go
func NewRouter(
    JWTService jwt.Service, 
    authHandler AuthHandler, 
    // ... existing handlers
    invitationHandler InvitationHandler,  // NEW
    storageBasePath string,
) *chi.Mux {
    // ... existing setup

    r.Route("/api/v1", func(r chi.Router) {
        // ... existing auth routes

        // Public invitation route
        r.Route("/invitations", func(r chi.Router) {
            r.Get("/{token}", invitationHandler.GetByToken)
        })

        // Authenticated but NO company required (for pending users)
        r.Group(func(r chi.Router) {
            r.Use(jwtauth.Verifier(JWTService.JWTAuth()))
            r.Use(middleware.AuthRequired(JWTService.JWTAuth()))
            // NO RequireCompany here!
            
            r.Route("/invitations", func(r chi.Router) {
                r.Get("/my", invitationHandler.ListMyInvitations)
                r.Post("/{token}/accept", invitationHandler.Accept)
            })
        })

        // Requires company
        r.Group(func(r chi.Router) {
            r.Use(jwtauth.Verifier(JWTService.JWTAuth()))
            r.Use(middleware.AuthRequired(JWTService.JWTAuth()))
            r.Use(middleware.RequireCompany)

            // ... existing routes

            r.Route("/employees", func(r chi.Router) {
                r.Get("/{id}", employeeHandler.GetEmployee)

                r.Group(func(r chi.Router) {
                    r.Use(middleware.RequireManager)
                    r.Get("/", employeeHandler.ListEmployees)
                    r.Get("/search", employeeHandler.SearchEmployees)
                    r.Post("/", employeeHandler.CreateEmployee)  // Auto creates invitation
                    r.Put("/{id}", employeeHandler.UpdateEmployee)
                    r.Delete("/{id}", employeeHandler.DeleteEmployee)
                    r.Post("/{id}/inactivate", employeeHandler.InactivateEmployee)
                    r.Post("/{id}/resend-invitation", employeeHandler.ResendInvitation)  // NEW
                    r.Delete("/{id}/invitation", employeeHandler.RevokeInvitation)       // NEW
                })

                r.Post("/{id}/avatar", employeeHandler.UploadAvatar)
            })
        })
    })
    return r
}
```

---

#### 17. Main Update
**File:** `cmd/api/main.go`

```go
import (
    // ... existing imports
    "github.com/cmlabs-hris/hris-backend-go/internal/pkg/email"
    invitationService "github.com/cmlabs-hris/hris-backend-go/internal/service/invitation"
)

func main() {
    // ... existing setup

    // Repositories
    invitationRepo := postgresql.NewInvitationRepository(db)
    // ... existing repos

    // Email Service
    emailService, err := email.NewEmailService(cfg.SMTP)
    if err != nil {
        log.Fatal("Failed to initialize email service:", err)
    }

    // Invitation Service
    invitationSvc := invitationService.NewInvitationService(
        db,
        invitationRepo,
        employeeRepo,
        userRepo,
        emailService,
        cfg.Invitation,
    )

    // Update Employee Service with invitation dependency
    employeeService := employeeService.NewEmployeeService(
        db,
        employeeRepo,
        invitationSvc,   // NEW
        companyRepo,     // NEW
        fileService,
    )

    // Handlers
    invitationHandler := appHTTP.NewInvitationHandler(invitationSvc)
    employeeHandler := appHTTP.NewEmployeeHandler(employeeService, invitationSvc)  // Updated
    // ... existing handlers

    // Router
    router := appHTTP.NewRouter(
        JWTService,
        authHandler,
        // ... existing handlers
        invitationHandler,  // NEW
        cfg.Storage.BasePath,
    )

    // ... rest same
}
```

---

### User Repository Update Needed

**File:** `internal/domain/user/repository.go` (addition to interface)

```go
type UserRepository interface {
    // ... existing methods
    UpdateCompanyAndRole(ctx context.Context, userID, companyID string, role Role) error
}
```

**File:** `internal/repository/postgresql/user.go` (addition)

```go
func (r *userRepositoryImpl) UpdateCompanyAndRole(ctx context.Context, userID, companyID string, role user.Role) error {
    q := GetQuerier(ctx, r.db)
    
    query := `
        UPDATE users 
        SET company_id = $1, role = $2, updated_at = NOW()
        WHERE id = $3
    `
    
    _, err := q.Exec(ctx, query, companyID, string(role), userID)
    return err
}
```

---

### Summary Query Count

| Operation | Queries |
|-----------|---------|
| CreateEmployee | 6 total: check email pending, check code/NIK, get inviter, get company, INSERT employee, INSERT invitation |
| Accept Invitation | 5 total: get invitation, get user, get employee, UPDATE employee, UPDATE user, UPDATE invitation |
| List My Invitations | 1 query |
| Get By Token | 1 query |
| Resend | 2 queries: get pending, update token + send email |
| Revoke | 2 queries: get pending, update status |

---

### API Endpoints Summary

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| POST | `/employees` | Create employee + auto invitation | Manager+ |
| POST | `/employees/{id}/resend-invitation` | Resend invitation | Manager+ |
| DELETE | `/employees/{id}/invitation` | Revoke invitation | Manager+ |
| GET | `/invitations/my` | List my pending invitations | Authenticated |
| GET | `/invitations/{token}` | Get invitation details | Public |
| POST | `/invitations/{token}/accept` | Accept invitation | Authenticated |
