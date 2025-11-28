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

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		slog.Error("Failed to parse multipart form", "error", err)
		response.BadRequest(w, "Failed to parse form data", nil)
		return
	}

	dataJSON := r.FormValue("data")
	if dataJSON == "" {
		response.BadRequest(w, "Field 'data' is required", nil)
		return
	}

	var req company.CreateCompanyRequest
	if err := json.Unmarshal([]byte(dataJSON), &req); err != nil {
		slog.Error("Failed to unmarshal JSON data", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	file, fileHeader, err := r.FormFile("attachment")
	if err != nil && err != http.ErrMissingFile {
		// Error other than missing file
		slog.Error("Failed to get file from form", "error", err)
		response.BadRequest(w, "Invalid file upload", nil)
		return
	}

	req.File = file
	req.FileHeader = fileHeader
	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
	}

	company, err := c.companyService.Create(r.Context(), req)
	if err != nil {
		slog.Error("Failed to create company", "error", err)
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Company created successfully", company)
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
