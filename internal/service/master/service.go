package master

import (
	"context"
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/branch"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/grade"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/position"
)

type MasterService interface {
	// Branch operations
	CreateBranch(ctx context.Context, req branch.CreateBranchRequest) (branch.BranchResponse, error)
	GetBranch(ctx context.Context, id string) (branch.BranchResponse, error)
	ListBranches(ctx context.Context, companyID string) ([]branch.BranchResponse, error)
	UpdateBranch(ctx context.Context, req branch.UpdateBranchRequest) error
	DeleteBranch(ctx context.Context, id string) error

	// Grade operations
	CreateGrade(ctx context.Context, companyID string, req grade.CreateGradeRequest) (grade.GradeResponse, error)
	GetGrade(ctx context.Context, id string) (grade.GradeResponse, error)
	ListGrades(ctx context.Context, companyID string) ([]grade.GradeResponse, error)
	UpdateGrade(ctx context.Context, req grade.UpdateGradeRequest) error
	DeleteGrade(ctx context.Context, id string) error

	// Position operations
	CreatePosition(ctx context.Context, companyID string, req position.CreatePositionRequest) (position.PositionResponse, error)
	GetPosition(ctx context.Context, id string) (position.PositionResponse, error)
	ListPositions(ctx context.Context, companyID string) ([]position.PositionResponse, error)
	UpdatePosition(ctx context.Context, req position.UpdatePositionRequest) error
	DeletePosition(ctx context.Context, id string) error
}

type masterServiceImpl struct {
	branchRepo   branch.BranchRepository
	gradeRepo    grade.GradeRepository
	positionRepo position.PositionRepository
}

func NewMasterService(
	branchRepo branch.BranchRepository,
	gradeRepo grade.GradeRepository,
	positionRepo position.PositionRepository,
) MasterService {
	return &masterServiceImpl{
		branchRepo:   branchRepo,
		gradeRepo:    gradeRepo,
		positionRepo: positionRepo,
	}
}

// ==================== BRANCH OPERATIONS ====================

func (s *masterServiceImpl) CreateBranch(ctx context.Context, req branch.CreateBranchRequest) (branch.BranchResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return branch.BranchResponse{}, err
	}

	// Create entity
	entity := branch.Branch{
		CompanyID: req.CompanyID,
		Name:      req.Name,
		Address:   req.Address,
	}

	// Save to database
	created, err := s.branchRepo.Create(ctx, entity)
	if err != nil {
		return branch.BranchResponse{}, fmt.Errorf("failed to create branch: %w", err)
	}

	// Map to response
	return branch.BranchResponse{
		ID:        created.ID,
		CompanyID: created.CompanyID,
		Name:      created.Name,
		Address:   created.Address,
	}, nil
}

func (s *masterServiceImpl) GetBranch(ctx context.Context, id string) (branch.BranchResponse, error) {
	entity, err := s.branchRepo.GetByID(ctx, id)
	if err != nil {
		return branch.BranchResponse{}, err
	}

	return branch.BranchResponse{
		ID:        entity.ID,
		CompanyID: entity.CompanyID,
		Name:      entity.Name,
		Address:   entity.Address,
	}, nil
}

func (s *masterServiceImpl) ListBranches(ctx context.Context, companyID string) ([]branch.BranchResponse, error) {
	branches, err := s.branchRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	var responses []branch.BranchResponse
	for _, b := range branches {
		responses = append(responses, branch.BranchResponse{
			ID:        b.ID,
			CompanyID: b.CompanyID,
			Name:      b.Name,
			Address:   b.Address,
		})
	}

	return responses, nil
}

func (s *masterServiceImpl) UpdateBranch(ctx context.Context, req branch.UpdateBranchRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// Update in database
	return s.branchRepo.Update(ctx, req)
}

func (s *masterServiceImpl) DeleteBranch(ctx context.Context, id string) error {
	return s.branchRepo.Delete(ctx, id)
}

// ==================== GRADE OPERATIONS ====================

func (s *masterServiceImpl) CreateGrade(ctx context.Context, companyID string, req grade.CreateGradeRequest) (grade.GradeResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return grade.GradeResponse{}, err
	}

	// Create entity
	entity := grade.Grade{
		CompanyID: companyID,
		Name:      req.Name,
	}

	// Save to database
	created, err := s.gradeRepo.Create(ctx, entity)
	if err != nil {
		return grade.GradeResponse{}, fmt.Errorf("failed to create grade: %w", err)
	}

	// Map to response
	return grade.GradeResponse{
		ID:   created.ID,
		Name: created.Name,
	}, nil
}

func (s *masterServiceImpl) GetGrade(ctx context.Context, id string) (grade.GradeResponse, error) {
	entity, err := s.gradeRepo.GetByID(ctx, id)
	if err != nil {
		return grade.GradeResponse{}, err
	}

	return grade.GradeResponse{
		ID:   entity.ID,
		Name: entity.Name,
	}, nil
}

func (s *masterServiceImpl) ListGrades(ctx context.Context, companyID string) ([]grade.GradeResponse, error) {
	grades, err := s.gradeRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	var responses []grade.GradeResponse
	for _, g := range grades {
		responses = append(responses, grade.GradeResponse{
			ID:   g.ID,
			Name: g.Name,
		})
	}

	return responses, nil
}

func (s *masterServiceImpl) UpdateGrade(ctx context.Context, req grade.UpdateGradeRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// Update in database
	return s.gradeRepo.Update(ctx, req)
}

func (s *masterServiceImpl) DeleteGrade(ctx context.Context, id string) error {
	return s.gradeRepo.Delete(ctx, id)
}

// ==================== POSITION OPERATIONS ====================

func (s *masterServiceImpl) CreatePosition(ctx context.Context, companyID string, req position.CreatePositionRequest) (position.PositionResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return position.PositionResponse{}, err
	}

	// Create entity
	entity := position.Position{
		CompanyID: companyID,
		Name:      req.Name,
	}

	// Save to database
	created, err := s.positionRepo.Create(ctx, entity)
	if err != nil {
		return position.PositionResponse{}, fmt.Errorf("failed to create position: %w", err)
	}

	// Map to response
	return position.PositionResponse{
		ID:        created.ID,
		CompanyID: created.CompanyID,
		Name:      created.Name,
	}, nil
}

func (s *masterServiceImpl) GetPosition(ctx context.Context, id string) (position.PositionResponse, error) {
	entity, err := s.positionRepo.GetByID(ctx, id)
	if err != nil {
		return position.PositionResponse{}, err
	}

	return position.PositionResponse{
		ID:        entity.ID,
		CompanyID: entity.CompanyID,
		Name:      entity.Name,
	}, nil
}

func (s *masterServiceImpl) ListPositions(ctx context.Context, companyID string) ([]position.PositionResponse, error) {
	positions, err := s.positionRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	var responses []position.PositionResponse
	for _, p := range positions {
		responses = append(responses, position.PositionResponse{
			ID:        p.ID,
			CompanyID: p.CompanyID,
			Name:      p.Name,
		})
	}

	return responses, nil
}

func (s *masterServiceImpl) UpdatePosition(ctx context.Context, req position.UpdatePositionRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// Update in database
	return s.positionRepo.Update(ctx, req)
}

func (s *masterServiceImpl) DeletePosition(ctx context.Context, id string) error {
	return s.positionRepo.Delete(ctx, id)
}
