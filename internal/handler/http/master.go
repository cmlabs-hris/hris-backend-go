package http

import (
	"encoding/json"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/branch"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/grade"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/position"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/master"
	"github.com/go-chi/chi/v5"
)

type MasterHandler interface {
	// Branch handlers
	CreateBranch(w http.ResponseWriter, r *http.Request)
	GetBranch(w http.ResponseWriter, r *http.Request)
	ListBranches(w http.ResponseWriter, r *http.Request)
	UpdateBranch(w http.ResponseWriter, r *http.Request)
	DeleteBranch(w http.ResponseWriter, r *http.Request)

	// Grade handlers
	CreateGrade(w http.ResponseWriter, r *http.Request)
	GetGrade(w http.ResponseWriter, r *http.Request)
	ListGrades(w http.ResponseWriter, r *http.Request)
	UpdateGrade(w http.ResponseWriter, r *http.Request)
	DeleteGrade(w http.ResponseWriter, r *http.Request)

	// Position handlers
	CreatePosition(w http.ResponseWriter, r *http.Request)
	GetPosition(w http.ResponseWriter, r *http.Request)
	ListPositions(w http.ResponseWriter, r *http.Request)
	UpdatePosition(w http.ResponseWriter, r *http.Request)
	DeletePosition(w http.ResponseWriter, r *http.Request)
}

type masterHandlerImpl struct {
	masterService master.MasterService
}

func NewMasterHandler(masterService master.MasterService) MasterHandler {
	return &masterHandlerImpl{
		masterService: masterService,
	}
}

// ==================== BRANCH HANDLERS ====================

func (h *masterHandlerImpl) CreateBranch(w http.ResponseWriter, r *http.Request) {
	var req branch.CreateBranchRequest

	// Get company ID from context
	companyID := r.Context().Value("company_id").(string)
	req.CompanyID = companyID

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	// Create branch
	result, err := h.masterService.CreateBranch(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Branch created successfully", result)
}

func (h *masterHandlerImpl) GetBranch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.masterService.GetBranch(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *masterHandlerImpl) ListBranches(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value("company_id").(string)

	results, err := h.masterService.ListBranches(r.Context(), companyID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

func (h *masterHandlerImpl) UpdateBranch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req branch.UpdateBranchRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := h.masterService.UpdateBranch(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Branch updated successfully"})
}

func (h *masterHandlerImpl) DeleteBranch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.masterService.DeleteBranch(r.Context(), id); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Branch deleted successfully"})
}

// ==================== GRADE HANDLERS ====================

func (h *masterHandlerImpl) CreateGrade(w http.ResponseWriter, r *http.Request) {
	var req grade.CreateGradeRequest

	companyID := r.Context().Value("company_id").(string)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	result, err := h.masterService.CreateGrade(r.Context(), companyID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Grade created successfully", result)
}

func (h *masterHandlerImpl) GetGrade(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.masterService.GetGrade(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *masterHandlerImpl) ListGrades(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value("company_id").(string)

	results, err := h.masterService.ListGrades(r.Context(), companyID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

func (h *masterHandlerImpl) UpdateGrade(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req grade.UpdateGradeRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := h.masterService.UpdateGrade(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Grade updated successfully"})
}

func (h *masterHandlerImpl) DeleteGrade(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.masterService.DeleteGrade(r.Context(), id); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Grade deleted successfully"})
}

// ==================== POSITION HANDLERS ====================

func (h *masterHandlerImpl) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var req position.CreatePositionRequest

	companyID := r.Context().Value("company_id").(string)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	result, err := h.masterService.CreatePosition(r.Context(), companyID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "Position created successfully", result)
}

func (h *masterHandlerImpl) GetPosition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.masterService.GetPosition(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

func (h *masterHandlerImpl) ListPositions(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value("company_id").(string)

	results, err := h.masterService.ListPositions(r.Context(), companyID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

func (h *masterHandlerImpl) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req position.UpdatePositionRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := h.masterService.UpdatePosition(r.Context(), req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Position updated successfully"})
}

func (h *masterHandlerImpl) DeletePosition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.masterService.DeletePosition(r.Context(), id); err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Position deleted successfully"})
}
