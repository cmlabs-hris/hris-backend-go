package http

import (
	"net/http"

	empDashboard "github.com/cmlabs-hris/hris-backend-go/internal/domain/employee_dashboard"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
)

type EmployeeDashboardHandler interface {
	// GetDashboard returns combined employee dashboard data
	GetDashboard(w http.ResponseWriter, r *http.Request)
	// GetWorkStats returns work stats for a date range
	GetWorkStats(w http.ResponseWriter, r *http.Request)
	// GetAttendanceSummary returns attendance summary for a month
	GetAttendanceSummary(w http.ResponseWriter, r *http.Request)
	// GetLeaveSummary returns leave quota summary for a year
	GetLeaveSummary(w http.ResponseWriter, r *http.Request)
	// GetWorkHoursChart returns daily work hours for a specific week
	GetWorkHoursChart(w http.ResponseWriter, r *http.Request)
}

type employeeDashboardHandlerImpl struct {
	service empDashboard.EmployeeDashboardService
}

func NewEmployeeDashboardHandler(service empDashboard.EmployeeDashboardService) EmployeeDashboardHandler {
	return &employeeDashboardHandlerImpl{service: service}
}

// GetDashboard handles GET /my-dashboard
// Returns combined dashboard with current month's attendance summary,
// current year's leave summary, and current week's work hours chart
func (h *employeeDashboardHandlerImpl) GetDashboard(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetDashboard(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetWorkStats handles GET /my-dashboard/work-stats
// Query params:
//   - start_date: YYYY-MM-DD (default: first day of current month)
//   - end_date: YYYY-MM-DD (default: last day of current month)
func (h *employeeDashboardHandlerImpl) GetWorkStats(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	result, err := h.service.GetWorkStats(r.Context(), startDate, endDate)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetAttendanceSummary handles GET /my-dashboard/attendance-summary
// Query params:
//   - month: YYYY-MM (default: current month)
func (h *employeeDashboardHandlerImpl) GetAttendanceSummary(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")

	result, err := h.service.GetAttendanceSummary(r.Context(), month)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetLeaveSummary handles GET /my-dashboard/leave-summary
// Query params:
//   - year: YYYY (default: current year)
func (h *employeeDashboardHandlerImpl) GetLeaveSummary(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")

	result, err := h.service.GetLeaveSummary(r.Context(), year)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetWorkHoursChart handles GET /my-dashboard/work-hours-chart
// Query params:
//   - week: 1, 2, 3, 4, 5 (default: current week of month)
func (h *employeeDashboardHandlerImpl) GetWorkHoursChart(w http.ResponseWriter, r *http.Request) {
	week := r.URL.Query().Get("week")

	result, err := h.service.GetWorkHoursChart(r.Context(), week)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}
