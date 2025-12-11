package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/payroll"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/chi/v5"
)

type PayrollHandler interface {
	// Settings
	GetSettings(w http.ResponseWriter, r *http.Request)
	UpdateSettings(w http.ResponseWriter, r *http.Request)

	// Components
	CreateComponent(w http.ResponseWriter, r *http.Request)
	GetComponent(w http.ResponseWriter, r *http.Request)
	ListComponents(w http.ResponseWriter, r *http.Request)
	UpdateComponent(w http.ResponseWriter, r *http.Request)
	DeleteComponent(w http.ResponseWriter, r *http.Request)

	// Employee Components
	AssignComponent(w http.ResponseWriter, r *http.Request)
	GetEmployeeComponents(w http.ResponseWriter, r *http.Request)
	UpdateEmployeeComponent(w http.ResponseWriter, r *http.Request)
	RemoveEmployeeComponent(w http.ResponseWriter, r *http.Request)

	// Payroll Records
	GeneratePayroll(w http.ResponseWriter, r *http.Request)
	GetPayrollRecord(w http.ResponseWriter, r *http.Request)
	ListPayrollRecords(w http.ResponseWriter, r *http.Request)
	UpdatePayrollRecord(w http.ResponseWriter, r *http.Request)
	FinalizePayroll(w http.ResponseWriter, r *http.Request)
	DeletePayrollRecord(w http.ResponseWriter, r *http.Request)

	// Summary
	GetPayrollSummary(w http.ResponseWriter, r *http.Request)
}

type payrollHandlerImpl struct {
	payrollService payroll.PayrollService
}

func NewPayrollHandler(payrollService payroll.PayrollService) PayrollHandler {
	return &payrollHandlerImpl{payrollService: payrollService}
}

// ========== SETTINGS ==========

func (h *payrollHandlerImpl) GetSettings(w http.ResponseWriter, r *http.Request) {
	result, err := h.payrollService.GetSettings(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *payrollHandlerImpl) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req payroll.UpdatePayrollSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	result, err := h.payrollService.UpdateSettings(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// ========== COMPONENTS ==========

func (h *payrollHandlerImpl) CreateComponent(w http.ResponseWriter, r *http.Request) {
	var req payroll.CreatePayrollComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	result, err := h.payrollService.CreateComponent(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Payroll component created", result)
}

func (h *payrollHandlerImpl) GetComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Component ID is required", nil)
		return
	}

	result, err := h.payrollService.GetComponent(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *payrollHandlerImpl) ListComponents(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	result, err := h.payrollService.ListComponents(r.Context(), activeOnly)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *payrollHandlerImpl) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Component ID is required", nil)
		return
	}

	var req payroll.UpdatePayrollComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}
	req.ID = id

	if err := h.payrollService.UpdateComponent(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, nil)
}

func (h *payrollHandlerImpl) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Component ID is required", nil)
		return
	}

	if err := h.payrollService.DeleteComponent(r.Context(), id); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Payroll component deleted successfully", nil)
}

// ========== EMPLOYEE COMPONENTS ==========

func (h *payrollHandlerImpl) AssignComponent(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeId")
	if employeeID == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	var req payroll.AssignComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}
	req.EmployeeID = employeeID

	result, err := h.payrollService.AssignComponentToEmployee(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Component assigned to employee", result)
}

func (h *payrollHandlerImpl) GetEmployeeComponents(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeId")
	if employeeID == "" {
		response.BadRequest(w, "Employee ID is required", nil)
		return
	}

	result, err := h.payrollService.GetEmployeeComponents(r.Context(), employeeID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *payrollHandlerImpl) UpdateEmployeeComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee component ID is required", nil)
		return
	}

	var req payroll.UpdateEmployeeComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}
	req.ID = id

	if err := h.payrollService.UpdateEmployeeComponent(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, nil)
}

func (h *payrollHandlerImpl) RemoveEmployeeComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Employee component ID is required", nil)
		return
	}

	if err := h.payrollService.RemoveEmployeeComponent(r.Context(), id); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Employee component removed successfully", nil)
}

// ========== PAYROLL RECORDS ==========

func (h *payrollHandlerImpl) GeneratePayroll(w http.ResponseWriter, r *http.Request) {
	var req payroll.GeneratePayrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	result, err := h.payrollService.GeneratePayroll(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Payroll generated", result)
}

func (h *payrollHandlerImpl) GetPayrollRecord(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Record ID is required", nil)
		return
	}

	result, err := h.payrollService.GetPayrollRecord(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *payrollHandlerImpl) ListPayrollRecords(w http.ResponseWriter, r *http.Request) {
	filter := payroll.PayrollFilter{
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if monthStr := r.URL.Query().Get("period_month"); monthStr != "" {
		if month, err := strconv.Atoi(monthStr); err == nil {
			filter.PeriodMonth = &month
		}
	}
	if yearStr := r.URL.Query().Get("period_year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filter.PeriodYear = &year
		}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		filter.EmployeeID = &employeeID
	}
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	result, err := h.payrollService.ListPayrollRecords(r.Context(), filter)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *payrollHandlerImpl) UpdatePayrollRecord(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Record ID is required", nil)
		return
	}

	var req payroll.UpdatePayrollRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}
	req.ID = id

	result, err := h.payrollService.UpdatePayrollRecord(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *payrollHandlerImpl) FinalizePayroll(w http.ResponseWriter, r *http.Request) {
	var req payroll.FinalizePayrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := h.payrollService.FinalizePayroll(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, nil)
}

func (h *payrollHandlerImpl) DeletePayrollRecord(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "Record ID is required", nil)
		return
	}

	if err := h.payrollService.DeletePayrollRecord(r.Context(), id); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Payroll record deleted successfully", nil)
}

// ========== SUMMARY ==========

func (h *payrollHandlerImpl) GetPayrollSummary(w http.ResponseWriter, r *http.Request) {
	monthStr := r.URL.Query().Get("period_month")
	yearStr := r.URL.Query().Get("period_year")

	if monthStr == "" || yearStr == "" {
		response.BadRequest(w, "period_month and period_year are required", nil)
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(w, "Invalid period_month", nil)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 {
		response.BadRequest(w, "Invalid period_year", nil)
		return
	}

	result, err := h.payrollService.GetPayrollSummary(r.Context(), month, year)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}
