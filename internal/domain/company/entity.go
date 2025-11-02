package company

import "time"

type Company struct {
	ID        string
	Name      string
	Username  string
	Address   *string
	LogoURL   *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
