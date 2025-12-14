package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/chi/v5"
)

type ScheduleHandler interface {
	// Work Schedule
	CreateWorkSchedule(w http.ResponseWriter, r *http.Request)
	GetWorkSchedule(w http.ResponseWriter, r *http.Request)
	ListWorkSchedules(w http.ResponseWriter, r *http.Request)
	UpdateWorkSchedule(w http.ResponseWriter, r *http.Request)
	DeleteWorkSchedule(w http.ResponseWriter, r *http.Request)

	// Work Schedule Time
	CreateWorkScheduleTime(w http.ResponseWriter, r *http.Request)
	GetWorkScheduleTime(w http.ResponseWriter, r *http.Request)
	UpdateWorkScheduleTime(w http.ResponseWriter, r *http.Request)
	DeleteWorkScheduleTime(w http.ResponseWriter, r *http.Request)

	// Work Schedule Location
	CreateWorkScheduleLocation(w http.ResponseWriter, r *http.Request)
	GetWorkScheduleLocation(w http.ResponseWriter, r *http.Request)
	UpdateWorkScheduleLocation(w http.ResponseWriter, r *http.Request)
	DeleteWorkScheduleLocation(w http.ResponseWriter, r *http.Request)

	// Employee Schedule Assignment
	CreateEmployeeScheduleAssignment(w http.ResponseWriter, r *http.Request)
	GetEmployeeScheduleAssignment(w http.ResponseWriter, r *http.Request)
	ListEmployeeScheduleAssignments(w http.ResponseWriter, r *http.Request)
	UpdateEmployeeScheduleAssignment(w http.ResponseWriter, r *http.Request)
	DeleteEmployeeScheduleAssignment(w http.ResponseWriter, r *http.Request)
	GetActiveScheduleForEmployee(w http.ResponseWriter, r *http.Request)

	// Employee Schedule Timeline
	GetEmployeeScheduleTimeline(w http.ResponseWriter, r *http.Request)
	AssignSchedule(w http.ResponseWriter, r *http.Request)
}

type scheduleHandlerImpl struct {
	scheduleService schedule.ScheduleService
}

func NewScheduleHandler(scheduleService schedule.ScheduleService) ScheduleHandler {
	return &scheduleHandlerImpl{
		scheduleService: scheduleService,
	}
}

// AssignSchedule implements ScheduleHandler.
func (h *scheduleHandlerImpl) AssignSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "scheduleID")
	employeeID := chi.URLParam(r, "employeeID")

	var req schedule.AssignScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}
	req.EmployeeID = employeeID
	req.WorkScheduleID = scheduleID

	result, err := h.scheduleService.AssignSchedule(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Schedule assigned successfully", result)
}

// ==================== WORK SCHEDULE HANDLERS ====================

func (h *scheduleHandlerImpl) CreateWorkSchedule(w http.ResponseWriter, r *http.Request) {
	var req schedule.CreateWorkScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.scheduleService.CreateWorkSchedule(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Work schedule created successfully", result)
}

func (h *scheduleHandlerImpl) GetWorkSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.scheduleService.GetWorkSchedule(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *scheduleHandlerImpl) ListWorkSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	filter := schedule.WorkScheduleFilter{}

	// Name filter
	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}

	// Type filter
	if scheduleType := r.URL.Query().Get("type"); scheduleType != "" {
		filter.Type = &scheduleType
	}

	// Check if requesting all records (no pagination)
	if allStr := r.URL.Query().Get("all"); allStr == "true" {
		filter.All = true
	} else {
		// Pagination (only when not fetching all)
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
	}

	// Sorting
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	// Get data from service
	results, err := h.scheduleService.ListWorkSchedules(ctx, filter)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

func (h *scheduleHandlerImpl) UpdateWorkSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req schedule.UpdateWorkScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}
	req.ID = id

	err := h.scheduleService.UpdateWorkSchedule(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Work schedule updated successfully", nil)
}

func (h *scheduleHandlerImpl) DeleteWorkSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.scheduleService.DeleteWorkSchedule(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Work schedule deleted successfully", nil)
}

// ==================== WORK SCHEDULE TIME HANDLERS ====================

func (h *scheduleHandlerImpl) CreateWorkScheduleTime(w http.ResponseWriter, r *http.Request) {
	var req schedule.CreateWorkScheduleTimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.scheduleService.CreateWorkScheduleTime(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Work schedule time created successfully", result)
}

func (h *scheduleHandlerImpl) GetWorkScheduleTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.scheduleService.GetWorkScheduleTime(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *scheduleHandlerImpl) UpdateWorkScheduleTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req schedule.UpdateWorkScheduleTimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}
	req.ID = id

	err := h.scheduleService.UpdateWorkScheduleTime(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Work schedule time updated successfully", nil)
}

func (h *scheduleHandlerImpl) DeleteWorkScheduleTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.scheduleService.DeleteWorkScheduleTime(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Work schedule time deleted successfully", nil)
}

// ==================== WORK SCHEDULE LOCATION HANDLERS ====================

func (h *scheduleHandlerImpl) CreateWorkScheduleLocation(w http.ResponseWriter, r *http.Request) {
	var req schedule.CreateWorkScheduleLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.scheduleService.CreateWorkScheduleLocation(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Work schedule location created successfully", result)
}

func (h *scheduleHandlerImpl) GetWorkScheduleLocation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.scheduleService.GetWorkScheduleLocation(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *scheduleHandlerImpl) UpdateWorkScheduleLocation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req schedule.UpdateWorkScheduleLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}
	req.ID = id

	err := h.scheduleService.UpdateWorkScheduleLocation(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Work schedule location updated successfully", nil)
}

func (h *scheduleHandlerImpl) DeleteWorkScheduleLocation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.scheduleService.DeleteWorkScheduleLocation(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Work schedule location deleted successfully", nil)
}

// ==================== EMPLOYEE SCHEDULE ASSIGNMENT HANDLERS ====================

func (h *scheduleHandlerImpl) CreateEmployeeScheduleAssignment(w http.ResponseWriter, r *http.Request) {
	var req schedule.CreateEmployeeScheduleAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.scheduleService.CreateEmployeeScheduleAssignment(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Employee schedule assignment created successfully", result)
}

func (h *scheduleHandlerImpl) GetEmployeeScheduleAssignment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.scheduleService.GetEmployeeScheduleAssignment(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *scheduleHandlerImpl) ListEmployeeScheduleAssignments(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	if employeeID == "" {
		response.HandleError(w, schedule.ErrEmployeeIDRequired)
		return
	}

	results, err := h.scheduleService.ListEmployeeScheduleAssignments(r.Context(), employeeID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

func (h *scheduleHandlerImpl) UpdateEmployeeScheduleAssignment(w http.ResponseWriter, r *http.Request) {
	assignID := chi.URLParam(r, "assignID")
	employeeID := chi.URLParam(r, "employeeID")

	var req schedule.UpdateEmployeeScheduleAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, err)
		return
	}
	req.ID = assignID
	req.EmployeeID = employeeID

	err := h.scheduleService.UpdateEmployeeScheduleAssignment(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Employee schedule assignment updated successfully", nil)
}

func (h *scheduleHandlerImpl) DeleteEmployeeScheduleAssignment(w http.ResponseWriter, r *http.Request) {
	assignID := chi.URLParam(r, "assignID")

	err := h.scheduleService.DeleteEmployeeScheduleAssignment(r.Context(), assignID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Employee schedule assignment deleted successfully", nil)
}

func (h *scheduleHandlerImpl) GetActiveScheduleForEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	dateStr := r.URL.Query().Get("date") // YYYY-MM-DD format

	if employeeID == "" {
		response.HandleError(w, schedule.ErrEmployeeIDRequired)
		return
	}

	var date time.Time
	var err error
	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			response.HandleError(w, schedule.ErrInvalidDateFormat)
			return
		}
	} else {
		date = time.Now()
	}

	result, err := h.scheduleService.GetActiveScheduleForEmployee(r.Context(), employeeID, date)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// ==================== EMPLOYEE SCHEDULE TIMELINE HANDLERS ====================

func (h *scheduleHandlerImpl) GetEmployeeScheduleTimeline(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "id")

	// Parse query parameters
	params := r.URL.Query()

	// Pagination
	page := 1
	if p := params.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}

	limit := 10
	if l := params.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	filter := schedule.EmployeeScheduleTimelineFilter{
		Page:  page,
		Limit: limit,
	}

	// Get timeline data from service
	result, err := h.scheduleService.GetEmployeeScheduleTimeline(r.Context(), employeeID, filter)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}
