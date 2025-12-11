package http

import (
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/dashboard"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
)

type DashboardHandler interface {
	// GetDashboard returns combined dashboard data
	GetDashboard(w http.ResponseWriter, r *http.Request)
	// GetEmployeeCurrentNumber returns new/active/resign counts
	GetEmployeeCurrentNumber(w http.ResponseWriter, r *http.Request)
	// GetEmployeeStatusStats returns employee counts by type
	GetEmployeeStatusStats(w http.ResponseWriter, r *http.Request)
	// GetMonthlyAttendance returns attendance stats for a month
	GetMonthlyAttendance(w http.ResponseWriter, r *http.Request)
	// GetDailyAttendanceStats returns attendance stats for a day
	GetDailyAttendanceStats(w http.ResponseWriter, r *http.Request)
}

type dashboardHandlerImpl struct {
	dashboardService dashboard.DashboardService
}

func NewDashboardHandler(dashboardService dashboard.DashboardService) DashboardHandler {
	return &dashboardHandlerImpl{dashboardService: dashboardService}
}

// GetDashboard handles GET /dashboard
func (h *dashboardHandlerImpl) GetDashboard(w http.ResponseWriter, r *http.Request) {
	result, err := h.dashboardService.GetDashboard(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetEmployeeCurrentNumber handles GET /dashboard/employee-current-number
func (h *dashboardHandlerImpl) GetEmployeeCurrentNumber(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month") // format: YYYY-MM, default: current month

	result, err := h.dashboardService.GetEmployeeCurrentNumber(r.Context(), month)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetEmployeeStatusStats handles GET /dashboard/employee-status-stats
func (h *dashboardHandlerImpl) GetEmployeeStatusStats(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month") // format: YYYY-MM, default: current month

	result, err := h.dashboardService.GetEmployeeStatusStats(r.Context(), month)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetMonthlyAttendance handles GET /dashboard/monthly-attendance
func (h *dashboardHandlerImpl) GetMonthlyAttendance(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month") // format: YYYY-MM, default: current month

	result, err := h.dashboardService.GetMonthlyAttendance(r.Context(), month)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetDailyAttendanceStats handles GET /dashboard/daily-attendance-stats
func (h *dashboardHandlerImpl) GetDailyAttendanceStats(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date") // format: YYYY-MM-DD, default: today

	result, err := h.dashboardService.GetDailyAttendanceStats(r.Context(), date)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}
