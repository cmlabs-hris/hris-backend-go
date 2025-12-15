package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type LeaveHandler interface {
	CreateType(w http.ResponseWriter, r *http.Request)
	UpdateType(w http.ResponseWriter, r *http.Request)
	ListTypes(w http.ResponseWriter, r *http.Request)
	DeleteType(w http.ResponseWriter, r *http.Request)

	SetQuota(w http.ResponseWriter, r *http.Request)
	ListQuota(w http.ResponseWriter, r *http.Request)
	AdjustQuota(w http.ResponseWriter, r *http.Request)
	GetMyQuota(w http.ResponseWriter, r *http.Request)
	GetQuota(w http.ResponseWriter, r *http.Request)

	ListRequests(w http.ResponseWriter, r *http.Request)
	GetMyRequests(w http.ResponseWriter, r *http.Request)
	GetRequest(w http.ResponseWriter, r *http.Request)
	CreateRequest(w http.ResponseWriter, r *http.Request)
	ApproveRequest(w http.ResponseWriter, r *http.Request)
	RejectRequest(w http.ResponseWriter, r *http.Request)
}

type LeaveHandlerImpl struct {
	leaveService leave.LeaveService
	fileService  file.FileService
}

// GetQuota implements LeaveHandler.
func (l *LeaveHandlerImpl) GetQuota(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse quota ID from URL
	quotaID := chi.URLParam(r, "id")
	if quotaID == "" {
		response.BadRequest(w, "Quota ID is required", nil)
		return
	}

	// Call service to get the quota
	leaveQuota, err := l.leaveService.GetLeaveQuota(ctx, quotaID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, leaveQuota)
}

// AdjustQuota implements LeaveHandler.
func (l *LeaveHandlerImpl) AdjustQuota(w http.ResponseWriter, r *http.Request) {
	var req leave.AdjustQuotaRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("AdjustQuota decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	if err := l.leaveService.AdjustLeaveQuota(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Leave quota adjusted successfully", nil)
}

// ApproveRequest implements LeaveHandler.
func (l *LeaveHandlerImpl) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	var req leave.ApproveRequestRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("ApproveRequest decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	if err := l.leaveService.ApproveLeaveRequest(r.Context(), req.RequestID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Leave request approved successfully", nil)
}

// CreateRequest implements LeaveHandler.
func (l *LeaveHandlerImpl) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var req leave.CreateLeaveRequestRequest

	// Get employee_id from JWT claims
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		slog.Error("Failed to get JWT claims", "error", err)
		response.Unauthorized(w, "Unauthorized")
		return
	}

	employeeID, ok := claims["employee_id"].(string)
	if !ok || employeeID == "" {
		slog.Error("employee_id not found in JWT claims")
		response.Forbidden(w, "Employee ID not found in token")
		return
	}

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

	if err := json.Unmarshal([]byte(dataJSON), &req); err != nil {
		slog.Error("Failed to unmarshal JSON data", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	// Set employee_id from JWT (override any value from request for security)
	req.EmployeeID = employeeID

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

	leaveRequest, err := l.leaveService.CreateLeaveRequest(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Leave request created successfully", leaveRequest)
}

// CreateType implements LeaveHandler.
func (l *LeaveHandlerImpl) CreateType(w http.ResponseWriter, r *http.Request) {
	var req leave.CreateLeaveTypeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("CreateType decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := req.Validate(); err != nil {
		slog.Error("CreateType validation error", "error", err)
		response.HandleError(w, err)
		return
	}

	leaveType, err := l.leaveService.CreateLeaveType(r.Context(), req)
	if err != nil {
		slog.Error("CreateType service error", "error", err)
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Leave type created successfully", leaveType)
}

// DeleteType implements LeaveHandler.
func (l *LeaveHandlerImpl) DeleteType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Leave type ID is required", nil)
		return
	}

	if err := l.leaveService.DeleteLeaveType(r.Context(), id); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Leave type deleted successfully", nil)
}

// GetMyQuota implements LeaveHandler.
func (l *LeaveHandlerImpl) GetMyQuota(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		response.Unauthorized(w, "Failed to extract claims from context")
		return
	}

	employeeID, ok := claims["employee_id"].(string)
	if !ok || employeeID == "" {
		response.Unauthorized(w, "employee_id claim is missing or invalid")
		return
	}

	leaveQuota, err := l.leaveService.GetMyQuota(r.Context(), employeeID, time.Now().Year())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, leaveQuota)
}

// GetMyRequests implements LeaveHandler.
func (l *LeaveHandlerImpl) GetMyRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		response.Unauthorized(w, "Failed to extract claims from context")
		return
	}

	employeeID, ok := claims["employee_id"].(string)
	if !ok || employeeID == "" {
		response.Unauthorized(w, "employee_id claim is missing or invalid")
		return
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		response.Unauthorized(w, "company_id claim is missing or invalid")
		return
	}

	// Parse query parameters
	filter := leave.MyLeaveRequestFilter{}

	// Leave type filter
	if leaveTypeID := r.URL.Query().Get("leave_type_id"); leaveTypeID != "" {
		filter.LeaveTypeID = &leaveTypeID
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	// Date range filters
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}

	// Pagination
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	// Sorting
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	// Call service
	leaveRequestResponse, err := l.leaveService.ListMyLeaveRequests(ctx, employeeID, companyID, filter)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, leaveRequestResponse)
}

// GetRequests implements LeaveHandler.
func (l *LeaveHandlerImpl) GetRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request ID from URL
	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		response.BadRequest(w, "Request ID is required", nil)
		return
	}

	// Call service to get the request
	leaveRequestResponse, err := l.leaveService.GetLeaveRequest(ctx, requestID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, leaveRequestResponse)
}

// ListQuota implements LeaveHandler.
func (l *LeaveHandlerImpl) ListQuota(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		response.Unauthorized(w, "Failed to extract claims from context")
		return
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		response.Unauthorized(w, "company_id claim is missing or invalid")
		return
	}

	leaveTypeResponse, err := l.leaveService.ListLeaveQuota(r.Context(), companyID)
	if err != nil {
		slog.Error("ListQuota service error", "error", err)
		response.HandleError(w, err)
		return
	}

	response.Success(w, leaveTypeResponse)
}

// ListRequests implements LeaveHandler.
func (l *LeaveHandlerImpl) ListRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		response.Unauthorized(w, "Failed to extract claims from context")
		return
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		response.Unauthorized(w, "company_id claim is missing or invalid")
		return
	}

	// Parse query parameters
	filter := leave.LeaveRequestFilter{}

	// Employee name filter
	if employeeName := r.URL.Query().Get("employee_name"); employeeName != "" {
		filter.EmployeeName = &employeeName
	}

	// Employee ID filter
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		filter.EmployeeID = &employeeID
	}

	// Leave type filter
	if leaveTypeID := r.URL.Query().Get("leave_type_id"); leaveTypeID != "" {
		filter.LeaveTypeID = &leaveTypeID
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	// Date range filter
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}

	// Pagination
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	filter.Page = page

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
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

	if err := filter.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	// Get data from service
	leaveRequestResponse, err := l.leaveService.ListLeaveRequest(ctx, companyID, filter)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, leaveRequestResponse)
}

// ListTypes implements LeaveHandler.
func (l *LeaveHandlerImpl) ListTypes(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		response.Unauthorized(w, "Failed to extract claims from context")
		return
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		response.Unauthorized(w, "company_id claim is missing or invalid")
		return
	}

	leaveTypes, err := l.leaveService.ListLeaveType(r.Context(), companyID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, leaveTypes)
}

// RejectRequest implements LeaveHandler.
func (l *LeaveHandlerImpl) RejectRequest(w http.ResponseWriter, r *http.Request) {
	var req leave.RejectRequestRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("RejectRequest decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	if err := l.leaveService.RejectLeaveRequest(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Leave request rejected successfully", nil)
}

// SetQuota implements LeaveHandler.
func (l *LeaveHandlerImpl) SetQuota(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// UpdateType implements LeaveHandler.
func (l *LeaveHandlerImpl) UpdateType(w http.ResponseWriter, r *http.Request) {
	var req leave.UpdateLeaveTypeRequest

	leaveTypeID := chi.URLParam(r, "id")
	if leaveTypeID == "" {
		response.BadRequest(w, "Leave type ID is required", nil)
	}

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("UpdateType decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	req.ID = leaveTypeID

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	if err := l.leaveService.UpdateLeaveType(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Leave type updated successfully", nil)
}

func NewLeaveHandler(leaveService leave.LeaveService, fileService file.FileService) LeaveHandler {
	return &LeaveHandlerImpl{
		leaveService: leaveService,
		fileService:  fileService,
	}
}
