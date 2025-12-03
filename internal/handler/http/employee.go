package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/invitation"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type EmployeeHandler interface {
	SearchEmployees(w http.ResponseWriter, r *http.Request)
	GetEmployee(w http.ResponseWriter, r *http.Request)
	CreateEmployee(w http.ResponseWriter, r *http.Request)
	UpdateEmployee(w http.ResponseWriter, r *http.Request)
	DeleteEmployee(w http.ResponseWriter, r *http.Request)
	ListEmployees(w http.ResponseWriter, r *http.Request)
	InactivateEmployee(w http.ResponseWriter, r *http.Request)
	UploadAvatar(w http.ResponseWriter, r *http.Request)
	ResendInvitation(w http.ResponseWriter, r *http.Request)
	RevokeInvitation(w http.ResponseWriter, r *http.Request)
}

type employeeHandlerImpl struct {
	employeeService   employee.EmployeeService
	invitationService invitation.InvitationService
}

func NewEmployeeHandler(employeeService employee.EmployeeService, invitationService invitation.InvitationService) EmployeeHandler {
	return &employeeHandlerImpl{
		employeeService:   employeeService,
		invitationService: invitationService,
	}
}

// SearchEmployees implements EmployeeHandler - autocomplete search
func (h *employeeHandlerImpl) SearchEmployees(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		query = r.URL.Query().Get("query")
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	req := employee.SearchEmployeeRequest{
		Query: query,
		Limit: limit,
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	results, err := h.employeeService.SearchEmployees(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

// GetEmployee implements EmployeeHandler
func (h *employeeHandlerImpl) GetEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	result, err := h.employeeService.GetEmployee(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// CreateEmployee implements EmployeeHandler
func (h *employeeHandlerImpl) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req employee.CreateEmployeeRequest

	contentType := r.Header.Get("Content-Type")

	// Check if it's multipart form (with file upload)
	if len(contentType) >= 19 && contentType[:19] == "multipart/form-data" {
		// Parse multipart form (max 10MB)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			slog.Error("Failed to parse multipart form", "error", err)
			response.BadRequest(w, "Failed to parse form data", nil)
			return
		}

		// Get JSON data from 'data' field
		dataJSON := r.FormValue("data")
		if dataJSON == "" {
			response.BadRequest(w, "Field 'data' is required", nil)
			return
		}

		// Unmarshal JSON data
		if err := json.Unmarshal([]byte(dataJSON), &req); err != nil {
			slog.Error("Failed to unmarshal JSON data", "error", err)
			response.BadRequest(w, "Invalid request format", nil)
			return
		}

		// Get optional avatar file
		file, fileHeader, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			req.File = file
			req.FileHeader = fileHeader
		}
	} else {
		// Regular JSON request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "Invalid request format", nil)
			return
		}
	}

	// Validate request
	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.employeeService.CreateEmployee(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Employee created successfully", result)
}

// UpdateEmployee implements EmployeeHandler
func (h *employeeHandlerImpl) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	var req employee.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}
	req.ID = id

	// Validate request
	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.employeeService.UpdateEmployee(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Employee updated successfully", result)
}

// DeleteEmployee implements EmployeeHandler
func (h *employeeHandlerImpl) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	err := h.employeeService.DeleteEmployee(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Employee deleted successfully", nil)
}

// ListEmployees implements EmployeeHandler
func (h *employeeHandlerImpl) ListEmployees(w http.ResponseWriter, r *http.Request) {
	filter := employee.EmployeeFilter{}

	// Search
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}

	// Filters
	if workScheduleID := r.URL.Query().Get("work_schedule_id"); workScheduleID != "" {
		filter.WorkScheduleID = &workScheduleID
	}
	if positionID := r.URL.Query().Get("position_id"); positionID != "" {
		filter.PositionID = &positionID
	}
	if gradeID := r.URL.Query().Get("grade_id"); gradeID != "" {
		filter.GradeID = &gradeID
	}
	if branchID := r.URL.Query().Get("branch_id"); branchID != "" {
		filter.BranchID = &branchID
	}
	if employmentType := r.URL.Query().Get("employment_type"); employmentType != "" {
		filter.EmploymentType = &employmentType
	}
	if employmentStatus := r.URL.Query().Get("employment_status"); employmentStatus != "" {
		filter.EmploymentStatus = &employmentStatus
	}
	if warningLetter := r.URL.Query().Get("warning_letter"); warningLetter != "" {
		filter.WarningLetter = &warningLetter
	}

	// Date ranges
	if hireDateFrom := r.URL.Query().Get("hire_date_from"); hireDateFrom != "" {
		filter.HireDateFrom = &hireDateFrom
	}
	if hireDateTo := r.URL.Query().Get("hire_date_to"); hireDateTo != "" {
		filter.HireDateTo = &hireDateTo
	}
	if resignationDateFrom := r.URL.Query().Get("resignation_date_from"); resignationDateFrom != "" {
		filter.ResignationDateFrom = &resignationDateFrom
	}
	if resignationDateTo := r.URL.Query().Get("resignation_date_to"); resignationDateTo != "" {
		filter.ResignationDateTo = &resignationDateTo
	}
	if dobFrom := r.URL.Query().Get("dob_from"); dobFrom != "" {
		filter.DOBFrom = &dobFrom
	}
	if dobTo := r.URL.Query().Get("dob_to"); dobTo != "" {
		filter.DOBTo = &dobTo
	}

	// Pagination
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	filter.Page = page

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if limitNum, err := strconv.Atoi(l); err == nil && limitNum > 0 {
			limit = limitNum
		}
	}
	filter.Limit = limit

	// Sorting
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	// Validate filter
	if err := filter.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	results, err := h.employeeService.ListEmployees(r.Context(), filter)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

// InactivateEmployee implements EmployeeHandler
func (h *employeeHandlerImpl) InactivateEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	var req employee.InactivateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}
	req.ID = id

	// Validate request
	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.employeeService.InactivateEmployee(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Employee inactivated successfully", result)
}

// UploadAvatar implements EmployeeHandler
func (h *employeeHandlerImpl) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	// Parse multipart form (max 5MB)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		slog.Error("Failed to parse multipart form", "error", err)
		response.BadRequest(w, "Failed to parse form data", nil)
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		if err == http.ErrMissingFile {
			response.BadRequest(w, "Avatar file is required", nil)
			return
		}
		slog.Error("Failed to get file from form", "error", err)
		response.BadRequest(w, "Invalid file upload", nil)
		return
	}
	defer file.Close()

	req := employee.UploadAvatarRequest{
		EmployeeID: id,
		File:       file,
		FileHeader: fileHeader,
	}

	// Validate request
	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.employeeService.UploadAvatar(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Avatar uploaded successfully", result)
}

// ResendInvitation implements EmployeeHandler
func (h *employeeHandlerImpl) ResendInvitation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	_, claims, _ := jwtauth.FromContext(r.Context())
	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		response.Unauthorized(w, "Company ID not found in token")
		return
	}

	err := h.invitationService.Resend(r.Context(), id, companyID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Invitation resent successfully", nil)
}

// RevokeInvitation implements EmployeeHandler
func (h *employeeHandlerImpl) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	_, claims, _ := jwtauth.FromContext(r.Context())
	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		response.Unauthorized(w, "Company ID not found in token")
		return
	}

	err := h.invitationService.Revoke(r.Context(), id, companyID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Invitation revoked successfully", nil)
}
