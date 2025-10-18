package user

import "time"

type User struct {
	ID                      string
	CompanyID               *string
	Email                   string
	PasswordHash            *string
	IsAdmin                 bool
	OAuthProvider           *string
	OAuthProviderID         *string
	EmailVerified           bool
	EmailVerificationToken  *string
	EmailVerificationSentAt *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
