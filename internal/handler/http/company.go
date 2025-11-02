package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	"github.com/go-chi/jwtauth/v5"
)

type CompanyHandler interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type CompanyHandlerImpl struct {
	jwtService     jwt.Service
	companyService company.CompanyService
	fileService    file.FileService
}

// Create implements CompanyHandler.
func (c *CompanyHandlerImpl) Create(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Delete implements CompanyHandler.
func (c *CompanyHandlerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// GetByID implements CompanyHandler.
func (c *CompanyHandlerImpl) GetByID(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// List implements CompanyHandler.
func (c *CompanyHandlerImpl) List(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Update implements CompanyHandler.
func (c *CompanyHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	var updateReq company.UpdateCompanyRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		slog.Error("Update company decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	// Validate DTO
	if err := updateReq.Validate(); err != nil {
		slog.Error("Update company validate error", "error", err)
		response.HandleError(w, err)
		return
	}
	// Get company_id from JWT
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		slog.Error("Failed to get JWT claims", "error", err)
		response.HandleError(w, auth.ErrInvalidToken)
		return
	}
	companyID, exist := claims["company_id"].(string)
	if companyID == "" || !exist {
		slog.Error("company_id not found in JWT claims", "claims", claims)
		response.HandleError(w, auth.ErrInvalidToken)
		return
	}

	// Call service
	err = c.companyService.Update(r.Context(), companyID, updateReq)
	if err != nil {
		slog.Error("Company update service error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Success response
	slog.Info("Update company successfully")
	response.SuccessWithMessage(w, "Company updated successfully", nil)
}

func NewCompanyHandler(jwtService jwt.Service, companyService company.CompanyService, fileService file.FileService) CompanyHandler {
	return &CompanyHandlerImpl{
		jwtService:     jwtService,
		companyService: companyService,
		fileService:    fileService,
	}
}
