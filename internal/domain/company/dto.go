package company

import (
	"mime/multipart"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

type CompanyResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"company_name"`
	Username  string     `json:"company_username"`
	Address   *string    `json:"company_address,omitempty"`
	LogoURL   *string    `json:"logo_url,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CreateCompanyRequest struct {
	Name          string                `json:"company_name"`
	Username      string                `json:"company_username"`
	Address       *string               `json:"company_address,omitempty"`
	AttachmentURL *string               `json:"-"`
	File          multipart.File        `json:"-"`
	FileHeader    *multipart.FileHeader `json:"-"`
}

func (r *CreateCompanyRequest) Validate() error {
	var errs validator.ValidationErrors

	// Company
	// if validator.IsEmpty(r.Name) {
	// 	errs = append(errs, validator.ValidationError{
	// 		Field:   "company_name",
	// 		Message: "company_name is required",
	// 	})
	// }
	// if len(r.Name) > 255 {
	// 	errs = append(errs, validator.ValidationError{
	// 		Field:   "company_name",
	// 		Message: "company_name must not exceed 255 characters",
	// 	})
	// }
	if validator.IsEmpty(r.Username) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username is required",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UpdateCompanyRequest struct {
	Name    *string `json:"company_name,omitempty"`
	Address *string `json:"company_address,omitempty"`
	LogoURL *string `json:"logo_url,omitempty"`
}

func (r *UpdateCompanyRequest) Validate() error {
	var errs validator.ValidationErrors

	// Company
	if r.Name != nil {
		if len(*r.Name) > 255 {
			errs = append(errs, validator.ValidationError{
				Field:   "company_name",
				Message: "company_name must not exceed 255 characters",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UploadCompanyLogoRequest struct {
	File       multipart.File        `json:"-"`
	FileHeader *multipart.FileHeader `json:"-"`
	CompanyID  string                `json:"-"`
}

func (r *UploadCompanyLogoRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.FileHeader == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "company logo photo is required",
		})
	}
	filename := r.FileHeader.Filename
	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		// Validate image format
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "invalid file type: only jpg, jpeg, png allowed",
		})
	}
	if r.FileHeader.Size > 10<<20 { // 10MB
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "company logo photo size must not exceed 10MB",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UploadCompanyLogoResponse struct {
	LogoURL string `json:"logo_url"`
}
