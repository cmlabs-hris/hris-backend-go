package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/chi/v5"
)

type AttendanceHandler interface {
	ClockIn(w http.ResponseWriter, r *http.Request)
	ClockOut(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	GetMyAttendance(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Approve(w http.ResponseWriter, r *http.Request)
	Reject(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type attendanceHandlerImpl struct {
	attendanceService attendance.AttendanceService
}

func NewAttendanceHandler(attendanceService attendance.AttendanceService) AttendanceHandler {
	return &attendanceHandlerImpl{
		attendanceService: attendanceService,
	}
}

// ClockIn implements AttendanceHandler.
func (h *attendanceHandlerImpl) ClockIn(w http.ResponseWriter, r *http.Request) {
	var req attendance.ClockInRequest

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

	// Get file from form
	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		if err == http.ErrMissingFile {
			response.BadRequest(w, "Attendance proof photo is required", nil)
			return
		}
		slog.Error("Failed to get file from form", "error", err)
		response.BadRequest(w, "Invalid file upload", nil)
		return
	}
	defer file.Close()

	// Attach file to request
	req.File = file
	req.FileHeader = fileHeader

	// Validate request
	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	// Call service
	result, err := h.attendanceService.ClockIn(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Clock in successful", result)
}

// ClockOut implements AttendanceHandler.
func (h *attendanceHandlerImpl) ClockOut(w http.ResponseWriter, r *http.Request) {
	var req attendance.ClockOutRequest

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

	// Get file from form
	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		if err == http.ErrMissingFile {
			response.BadRequest(w, "Attendance proof photo is required", nil)
			return
		}
		slog.Error("Failed to get file from form", "error", err)
		response.BadRequest(w, "Invalid file upload", nil)
		return
	}
	defer file.Close()

	// Attach file to request
	req.File = file
	req.FileHeader = fileHeader

	// Validate request
	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	// Call service
	result, err := h.attendanceService.ClockOut(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// List implements AttendanceHandler.
func (h *attendanceHandlerImpl) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	filter := attendance.AttendanceFilter{}

	// Employee ID filter
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		filter.EmployeeID = &employeeID
	}

	// Employee name filter
	if employeeName := r.URL.Query().Get("employee_name"); employeeName != "" {
		filter.EmployeeName = &employeeName
	}

	// Date filter
	if date := r.URL.Query().Get("date"); date != "" {
		filter.Date = &date
	}

	// Date range filters
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
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

	// Get data from service
	results, err := h.attendanceService.ListAttendance(ctx, filter)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

// GetMyAttendance implements AttendanceHandler.
func (h *attendanceHandlerImpl) GetMyAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	filter := attendance.MyAttendanceFilter{}

	// Date filter
	if date := r.URL.Query().Get("date"); date != "" {
		filter.Date = &date
	}

	// Date range filters
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
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

	// Get data from service
	results, err := h.attendanceService.GetMyAttendance(ctx, filter)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

// Update implements AttendanceHandler.
func (h *attendanceHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req attendance.UpdateAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}
	req.ID = id

	result, err := h.attendanceService.UpdateAttendance(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Attendance updated successfully", result)
}

// Get implements AttendanceHandler.
func (h *attendanceHandlerImpl) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.attendanceService.GetAttendance(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// Approve implements AttendanceHandler.
func (h *attendanceHandlerImpl) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req attendance.ApproveAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}
	req.ID = id

	result, err := h.attendanceService.ApproveAttendance(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Attendance approved successfully", result)
}

// Reject implements AttendanceHandler.
func (h *attendanceHandlerImpl) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req attendance.RejectAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}
	req.ID = id

	result, err := h.attendanceService.RejectAttendance(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Attendance rejected successfully", result)
}

// Delete implements AttendanceHandler.
func (h *attendanceHandlerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.attendanceService.DeleteAttendance(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Attendance deleted successfully", nil)
}
