package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/report"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
)

type ReportHandler interface {
	// Monthly Attendance Report
	GetMonthlyAttendanceReport(w http.ResponseWriter, r *http.Request)

	// Payroll Summary Report
	GetPayrollSummaryReport(w http.ResponseWriter, r *http.Request)

	// Leave Balance Report
	GetLeaveBalanceReport(w http.ResponseWriter, r *http.Request)

	// New Hire Report
	GetNewHireReport(w http.ResponseWriter, r *http.Request)
}

type reportHandlerImpl struct {
	reportService report.ReportService
}

func NewReportHandler(reportService report.ReportService) ReportHandler {
	return &reportHandlerImpl{
		reportService: reportService,
	}
}

// GetMonthlyAttendanceReport handles GET /reports/attendance
func (h *reportHandlerImpl) GetMonthlyAttendanceReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	monthStr := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		response.BadRequest(w, "invalid month parameter", nil)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.BadRequest(w, "invalid year parameter", nil)
		return
	}

	req := report.MonthlyAttendanceReportRequest{
		Month: month,
		Year:  year,
	}

	result, err := h.reportService.GenerateMonthlyAttendanceReport(ctx, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetPayrollSummaryReport handles GET /reports/payroll
func (h *reportHandlerImpl) GetPayrollSummaryReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	monthStr := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		response.BadRequest(w, "invalid month parameter", nil)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.BadRequest(w, "invalid year parameter", nil)
		return
	}

	req := report.PayrollSummaryReportRequest{
		Month: month,
		Year:  year,
	}

	result, err := h.reportService.GeneratePayrollSummaryReport(ctx, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetLeaveBalanceReport handles GET /reports/leave-balance
func (h *reportHandlerImpl) GetLeaveBalanceReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	yearStr := r.URL.Query().Get("year")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.BadRequest(w, "invalid year parameter", nil)
		return
	}

	req := report.LeaveBalanceReportRequest{
		Year: year,
	}

	result, err := h.reportService.GenerateLeaveBalanceReport(ctx, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// GetNewHireReport handles POST /reports/new-hire or GET with query params
func (h *reportHandlerImpl) GetNewHireReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req report.NewHireReportRequest

	// Support both GET with query params and POST with body
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body", nil)
			return
		}
	} else {
		req.StartDate = r.URL.Query().Get("start_date")
		req.EndDate = r.URL.Query().Get("end_date")
	}

	result, err := h.reportService.GenerateNewHireReport(ctx, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}
