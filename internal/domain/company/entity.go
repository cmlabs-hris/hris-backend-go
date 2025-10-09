package company

import "time"

type Company struct {
	ID        string
	Name      string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
