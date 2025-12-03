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

// EmailService defines the interface for sending emails
type EmailService interface {
	SendInvitation(to, employeeName, inviterName, companyName string, positionName *string, invitationLink, expiresAt string) error
}

type emailServiceImpl struct {
	cfg       config.SMTPConfig
	templates *template.Template
}

// NewEmailService creates a new email service instance
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

// SendInvitation sends an invitation email to the employee
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
	// Skip sending if SMTP is not configured
	if s.cfg.Host == "" {
		return nil
	}

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
