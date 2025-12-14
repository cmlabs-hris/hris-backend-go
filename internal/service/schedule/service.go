package schedule

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/notification"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type scheduleServiceImpl struct {
	db                         *database.DB
	workScheduleRepo           schedule.WorkScheduleRepository
	workScheduleTimeRepo       schedule.WorkScheduleTimeRepository
	workScheduleLocationRepo   schedule.WorkScheduleLocationRepository
	employeeScheduleAssignRepo schedule.EmployeeScheduleAssignmentRepository
	employeeRepo               employee.EmployeeRepository
	notificationService        notification.Service
}

// AssignSchedule implements schedule.ScheduleService.
func (s *scheduleServiceImpl) AssignSchedule(ctx context.Context, req schedule.AssignScheduleRequest) (schedule.AssignScheduleResponse, error) {
	if err := req.Validate(); err != nil {
		return schedule.AssignScheduleResponse{}, err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.AssignScheduleResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.AssignScheduleResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	var response = schedule.AssignScheduleResponse{
		EmployeeID:     req.EmployeeID,
		WorkScheduleID: req.WorkScheduleID,
		StartDate:      &req.StartDate,
		EndDate:        req.EndDate,
	}
	if req.EndDate == nil || *req.EndDate == "" {
		err := postgresql.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
			txCtx := context.WithValue(ctx, "tx", tx)
			if err := s.employeeRepo.UpdateSchedule(txCtx, req.EmployeeID, req.WorkScheduleID, companyID); err != nil {
				return schedule.ErrInvalidRequestData
			}

			startDate, _ := time.Parse("2006-01-02", req.StartDate)
			if err := s.employeeScheduleAssignRepo.DeleteFutureAssignments(txCtx, startDate, req.EmployeeID, companyID); err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("failed to delete future assignments: %w", err)
				}
			}
			return nil
		})
		if err != nil {
			return schedule.AssignScheduleResponse{}, err
		}
	} else {
		startDate, _ := time.Parse("2006-01-02", req.StartDate)
		endDate, _ := time.Parse("2006-01-02", *req.EndDate)
		assignment := schedule.EmployeeScheduleAssignment{
			EmployeeID:     req.EmployeeID,
			WorkScheduleID: req.WorkScheduleID,
			StartDate:      startDate,
			EndDate:        endDate,
		}
		created, err := s.employeeScheduleAssignRepo.Create(ctx, assignment, companyID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return schedule.AssignScheduleResponse{}, schedule.ErrInvalidRequestData
			}
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				// Check for exclusion violation (SQL state code '23P01')
				if pgErr.Code == "23P01" && pgErr.ConstraintName == "no_overlapping_schedules" {
					return schedule.AssignScheduleResponse{}, schedule.ErrOverlappingScheduleAssignment
				}
			}
			return schedule.AssignScheduleResponse{}, fmt.Errorf("failed to create employee schedule assignment: %w", err)
		}
		createdAt := created.CreatedAt.Format("2006-01-02T15:04:05Z")
		updatedAt := created.UpdatedAt.Format("2006-01-02T15:04:05Z")
		response.CreatedAt = &createdAt
		response.UpdatedAt = &updatedAt
	}

	// Notify employee about schedule assignment/update
	go s.notifyEmployeeOnScheduleUpdated(ctx, req.EmployeeID, req.WorkScheduleID, companyID, req.StartDate)

	return response, nil
}

// CreateEmployeeScheduleAssignment implements schedule.ScheduleService.
func (s *scheduleServiceImpl) CreateEmployeeScheduleAssignment(ctx context.Context, req schedule.CreateEmployeeScheduleAssignmentRequest) (schedule.EmployeeScheduleAssignmentResponse, error) {
	panic("unimplemented")
	// if err := req.Validate(); err != nil {
	// 	return schedule.EmployeeScheduleAssignmentResponse{}, err
	// }

	// startDate, _ := time.Parse("2006-01-02", req.StartDate)
	// var endDate *time.Time
	// if req.EndDate != nil {
	// 	t, _ := time.Parse("2006-01-02", *req.EndDate)
	// 	endDate = &t
	// }

	// if req.EndDate == nil {
	// 	if err := s.employeeRepo.Update(ctx, req.EmployeeID, employee.UpdateEmployeeRequest{WorkScheduleID: &req.WorkScheduleID}); err != nil {
	// 		return schedule.EmployeeScheduleAssignmentResponse{}, fmt.Errorf("failed to update employee work schedule: %w", err)
	// 	}
	// 	esa := schedule.EmployeeScheduleAssignment{
	// 		EmployeeID:     req.EmployeeID,
	// 		WorkScheduleID: req.WorkScheduleID,
	// 		StartDate:      startDate,
	// 		EndDate:        *endDate,
	// 	}
	// 	return s.mapEmployeeScheduleAssignmentToResponse(esa), nil

	// }

	// esa := schedule.EmployeeScheduleAssignment{
	// 	EmployeeID:     req.EmployeeID,
	// 	WorkScheduleID: req.WorkScheduleID,
	// 	StartDate:      startDate,
	// 	EndDate:        *endDate,
	// }

	// created, err := s.employeeScheduleAssignRepo.Create(ctx, esa, companyid)
	// if err != nil {
	// 	var pgErr *pgconn.PgError
	// 	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // exclusion_violation
	// 		return schedule.EmployeeScheduleAssignmentResponse{}, schedule.ErrOverlappingScheduleAssignment
	// 	}
	// 	return schedule.EmployeeScheduleAssignmentResponse{}, fmt.Errorf("failed to create employee schedule assignment: %w", err)
	// }

	// return s.mapEmployeeScheduleAssignmentToResponse(created), nil
}

// CreateWorkSchedule implements schedule.ScheduleService.
func (s *scheduleServiceImpl) CreateWorkSchedule(ctx context.Context, req schedule.CreateWorkScheduleRequest) (schedule.WorkScheduleResponse, error) {
	if err := req.Validate(); err != nil {
		return schedule.WorkScheduleResponse{}, err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	// Create work schedule
	ws := schedule.WorkSchedule{
		CompanyID:          companyID,
		Name:               req.Name,
		Type:               schedule.WorkArrangement(req.Type),
		GracePeriodMinutes: *req.GracePeriodMinutes,
	}

	createdSchedule, err := s.workScheduleRepo.Create(ctx, ws)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return schedule.WorkScheduleResponse{}, schedule.ErrWorkScheduleNameExists
		}
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to create work schedule: %w", err)
	}

	return schedule.WorkScheduleResponse{
		ID:                 createdSchedule.ID,
		CompanyID:          createdSchedule.CompanyID,
		Name:               createdSchedule.Name,
		Type:               string(createdSchedule.Type),
		GracePeriodMinutes: createdSchedule.GracePeriodMinutes,
		CreatedAt:          createdSchedule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          createdSchedule.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// CreateWorkScheduleLocation implements schedule.ScheduleService.
func (s *scheduleServiceImpl) CreateWorkScheduleLocation(ctx context.Context, req schedule.CreateWorkScheduleLocationRequest) (schedule.WorkScheduleLocationResponse, error) {
	if err := req.Validate(); err != nil {
		return schedule.WorkScheduleLocationResponse{}, err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.WorkScheduleLocationResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.WorkScheduleLocationResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	ws, err := s.workScheduleRepo.GetByID(ctx, req.WorkScheduleID, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.WorkScheduleLocationResponse{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.WorkScheduleLocationResponse{}, fmt.Errorf("failed to get work schedule: %w", err)
	}
	// Can only insert work schedule information if work schedule type is WFO
	if ws.Type != schedule.WorkArrangementWFO && ws.Type != schedule.WorkArrangementHybrid {
		return schedule.WorkScheduleLocationResponse{}, schedule.ErrInvalidWorkScheduleType
	}

	wsl := schedule.WorkScheduleLocation{
		WorkScheduleID: req.WorkScheduleID,
		LocationName:   req.LocationName,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		RadiusMeters:   req.RadiusMeters,
	}

	workScheduleLocation, err := s.workScheduleLocationRepo.Create(ctx, wsl, companyID)
	if err != nil {
		if errors.Is(err, schedule.ErrWorkScheduleNotFound) {
			return schedule.WorkScheduleLocationResponse{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.WorkScheduleLocationResponse{}, fmt.Errorf("failed to create work schedule location: %w", err)
	}

	return s.mapWorkScheduleLocationToResponse(workScheduleLocation), nil
}

// CreateWorkScheduleTime implements schedule.ScheduleService.
func (s *scheduleServiceImpl) CreateWorkScheduleTime(ctx context.Context, req schedule.CreateWorkScheduleTimeRequest) (schedule.WorkScheduleTimeResponse, error) {
	if err := req.Validate(); err != nil {
		return schedule.WorkScheduleTimeResponse{}, err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.WorkScheduleTimeResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.WorkScheduleTimeResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	clockIn, _ := time.Parse("15:04", req.ClockInTime)
	clockOut, _ := time.Parse("15:04", req.ClockOutTime)

	var breakStart, breakEnd *time.Time
	if req.BreakStartTime != nil {
		t, _ := time.Parse("15:04", *req.BreakStartTime)
		breakStart = &t
	}
	if req.BreakEndTime != nil {
		t, _ := time.Parse("15:04", *req.BreakEndTime)
		breakEnd = &t
	}

	wsData, err := s.workScheduleRepo.GetByID(ctx, req.WorkScheduleID, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.WorkScheduleTimeResponse{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.WorkScheduleTimeResponse{}, fmt.Errorf("failed to get work schedule: %w", err)
	}

	// location type validation
	if wsData.Type == schedule.WorkArrangementWFO && req.LocationType != string(schedule.WorkArrangementWFO) {
		return schedule.WorkScheduleTimeResponse{}, schedule.ErrMismatchedLocationType
	}

	if wsData.Type == schedule.WorkArrangementWFA && req.LocationType != string(schedule.WorkArrangementWFA) {
		return schedule.WorkScheduleTimeResponse{}, schedule.ErrMismatchedLocationType
	}

	wst := schedule.WorkScheduleTime{
		WorkScheduleID:    req.WorkScheduleID,
		DayOfWeek:         *req.DayOfWeek,
		ClockInTime:       clockIn,
		BreakStartTime:    breakStart,
		BreakEndTime:      breakEnd,
		ClockOutTime:      clockOut,
		LocationType:      schedule.WorkArrangement(req.LocationType),
		IsNextDayCheckout: *req.IsNextDayCheckout,
	}

	// Repository Create will verify work_schedule belongs to company via EXISTS subquery
	created, err := s.workScheduleTimeRepo.Create(ctx, wst, companyID)
	if err != nil {
		if errors.Is(err, schedule.ErrWorkScheduleNotFound) {
			return schedule.WorkScheduleTimeResponse{}, schedule.ErrWorkScheduleNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return schedule.WorkScheduleTimeResponse{}, schedule.ErrWorkScheduleTimeExists
		}
		return schedule.WorkScheduleTimeResponse{}, fmt.Errorf("failed to create work schedule time: %w", err)
	}

	return s.mapWorkScheduleTimeToResponse(created), nil
}

// DeleteEmployeeScheduleAssignment implements schedule.ScheduleService.
func (s *scheduleServiceImpl) DeleteEmployeeScheduleAssignment(ctx context.Context, id string) error {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	err = s.employeeScheduleAssignRepo.Delete(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.ErrEmployeeScheduleAssignmentNotFound
		}
		return fmt.Errorf("failed to delete employee schedule assignment: %w", err)
	}

	return nil
}

// DeleteWorkSchedule implements schedule.ScheduleService.
func (s *scheduleServiceImpl) DeleteWorkSchedule(ctx context.Context, id string) error {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	err = s.workScheduleRepo.SoftDelete(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, schedule.ErrWorkScheduleNotFound) {
			return schedule.ErrWorkScheduleNotFound
		}
		return fmt.Errorf("failed to delete work schedule: %w", err)
	}

	return nil
}

// DeleteWorkScheduleLocation implements schedule.ScheduleService.
func (s *scheduleServiceImpl) DeleteWorkScheduleLocation(ctx context.Context, id string) error {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	err = s.workScheduleLocationRepo.Delete(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, schedule.ErrWorkScheduleLocationNotFound) {
			return schedule.ErrWorkScheduleLocationNotFound
		}
		return fmt.Errorf("failed to delete work schedule location: %w", err)
	}

	return nil
}

// DeleteWorkScheduleTime implements schedule.ScheduleService.
func (s *scheduleServiceImpl) DeleteWorkScheduleTime(ctx context.Context, id string) error {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	err = s.workScheduleTimeRepo.Delete(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, schedule.ErrWorkScheduleTimeNotFound) {
			return schedule.ErrWorkScheduleTimeNotFound
		}
		return fmt.Errorf("failed to delete work schedule time: %w", err)
	}

	return nil
}

// GetActiveScheduleForEmployee implements schedule.ScheduleService.
func (s *scheduleServiceImpl) GetActiveScheduleForEmployee(ctx context.Context, employeeID string, date time.Time) (schedule.WorkScheduleResponse, error) {
	// TODO THIS SHOULD USE JOIN FOR BETTER OPTMIZATION
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	ws, err := s.employeeScheduleAssignRepo.GetActiveSchedule(ctx, employeeID, date)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.WorkScheduleResponse{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to get active schedule: %w", err)
	}

	// Get times
	times, err := s.workScheduleTimeRepo.GetByWorkScheduleID(ctx, ws.ID, companyID)
	if err != nil {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to get work schedule times: %w", err)
	}

	var timeResponses []schedule.WorkScheduleTimeResponse
	for _, t := range times {
		timeResponses = append(timeResponses, s.mapWorkScheduleTimeToResponse(t))
	}

	// Get locations
	locations, err := s.workScheduleLocationRepo.GetByWorkScheduleID(ctx, ws.ID, companyID)
	if err != nil {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to get work schedule locations: %w", err)
	}

	var locationResponses []schedule.WorkScheduleLocationResponse
	for _, loc := range locations {
		locationResponses = append(locationResponses, s.mapWorkScheduleLocationToResponse(loc))
	}

	return schedule.WorkScheduleResponse{
		ID:        ws.ID,
		CompanyID: ws.CompanyID,
		Name:      ws.Name,
		Type:      string(ws.Type),
		Times:     timeResponses,
		Locations: locationResponses,
	}, nil
}

// GetEmployeeScheduleAssignment implements schedule.ScheduleService.
func (s *scheduleServiceImpl) GetEmployeeScheduleAssignment(ctx context.Context, id string) (schedule.EmployeeScheduleAssignmentResponse, error) {
	panic("unimplemented")
}

// GetWorkSchedule implements schedule.ScheduleService.
func (s *scheduleServiceImpl) GetWorkSchedule(ctx context.Context, id string) (schedule.WorkScheduleResponse, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	ws, err := s.workScheduleRepo.GetByID(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.WorkScheduleResponse{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to get work schedule: %w", err)
	}

	// Get times
	times, err := s.workScheduleTimeRepo.GetByWorkScheduleID(ctx, ws.ID, companyID)
	if err != nil {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to get work schedule times: %w", err)
	}

	var timeResponses []schedule.WorkScheduleTimeResponse
	for _, t := range times {
		timeResponses = append(timeResponses, s.mapWorkScheduleTimeToResponse(t))
	}

	// Get locations
	locations, err := s.workScheduleLocationRepo.GetByWorkScheduleID(ctx, ws.ID, companyID)
	if err != nil {
		return schedule.WorkScheduleResponse{}, fmt.Errorf("failed to get work schedule locations: %w", err)
	}

	var locationResponses []schedule.WorkScheduleLocationResponse
	for _, loc := range locations {
		locationResponses = append(locationResponses, s.mapWorkScheduleLocationToResponse(loc))
	}

	return schedule.WorkScheduleResponse{
		ID:                 ws.ID,
		CompanyID:          ws.CompanyID,
		Name:               ws.Name,
		Type:               string(ws.Type),
		Times:              timeResponses,
		Locations:          locationResponses,
		GracePeriodMinutes: ws.GracePeriodMinutes,
		CreatedAt:          ws.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          ws.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetWorkScheduleLocation implements schedule.ScheduleService.
func (s *scheduleServiceImpl) GetWorkScheduleLocation(ctx context.Context, id string) (schedule.WorkScheduleLocationResponse, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.WorkScheduleLocationResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.WorkScheduleLocationResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	wsl, err := s.workScheduleLocationRepo.GetByID(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.WorkScheduleLocationResponse{}, schedule.ErrWorkScheduleLocationNotFound
		}
		return schedule.WorkScheduleLocationResponse{}, fmt.Errorf("failed to get work schedule location: %w", err)
	}

	return s.mapWorkScheduleLocationToResponse(wsl), nil
}

// GetWorkScheduleTime implements schedule.ScheduleService.
func (s *scheduleServiceImpl) GetWorkScheduleTime(ctx context.Context, id string) (schedule.WorkScheduleTimeResponse, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.WorkScheduleTimeResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.WorkScheduleTimeResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	wst, err := s.workScheduleTimeRepo.GetByID(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.WorkScheduleTimeResponse{}, schedule.ErrWorkScheduleTimeNotFound
		}
		return schedule.WorkScheduleTimeResponse{}, fmt.Errorf("failed to get work schedule time: %w", err)
	}

	return s.mapWorkScheduleTimeToResponse(wst), nil
}

// ListEmployeeScheduleAssignments implements schedule.ScheduleService.
func (s *scheduleServiceImpl) ListEmployeeScheduleAssignments(ctx context.Context, employeeID string) ([]schedule.EmployeeScheduleAssignmentResponse, error) {
	panic("unimplemented")
}

// ListWorkSchedules implements schedule.ScheduleService.
func (s *scheduleServiceImpl) ListWorkSchedules(ctx context.Context, filter schedule.WorkScheduleFilter) (schedule.ListWorkScheduleResponse, error) {
	// Validate filter
	if err := filter.Validate(); err != nil {
		return schedule.ListWorkScheduleResponse{}, err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.ListWorkScheduleResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.ListWorkScheduleResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	// Get schedules from repository
	workSchedules, totalCount, err := s.workScheduleRepo.GetByCompanyID(ctx, companyID, filter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.ListWorkScheduleResponse{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.ListWorkScheduleResponse{}, fmt.Errorf("failed to list work schedules: %w", err)
	}

	// Map to response
	var workScheduleResponses []schedule.WorkScheduleResponse
	for _, ws := range workSchedules {
		var timeResponse []schedule.WorkScheduleTimeResponse
		for _, wsTime := range ws.Times {
			timeResponse = append(timeResponse, s.mapWorkScheduleTimeToResponse(wsTime))
		}

		var locationResponse []schedule.WorkScheduleLocationResponse
		for _, wsLocation := range ws.Locations {
			locationResponse = append(locationResponse, s.mapWorkScheduleLocationToResponse(wsLocation))
		}
		workScheduleResponses = append(workScheduleResponses, schedule.WorkScheduleResponse{
			ID:                 ws.ID,
			CompanyID:          ws.CompanyID,
			Name:               ws.Name,
			Type:               string(ws.Type),
			GracePeriodMinutes: ws.GracePeriodMinutes,
			Times:              timeResponse,
			Locations:          locationResponse,
			CreatedAt:          ws.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          ws.UpdatedAt.Format(time.RFC3339),
		})

	}

	// Handle response for "all" vs paginated
	if filter.All {
		return schedule.ListWorkScheduleResponse{
			TotalCount:    totalCount,
			Page:          1,
			Limit:         int(totalCount),
			TotalPages:    1,
			Showing:       fmt.Sprintf("All %d results", totalCount),
			WorkSchedules: workScheduleResponses,
		}, nil
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.Limit)))

	// Calculate "showing" text
	start := (filter.Page-1)*filter.Limit + 1
	end := start + len(workScheduleResponses) - 1
	if end > int(totalCount) {
		end = int(totalCount)
	}

	showing := fmt.Sprintf("%d-%d of %d results", start, end, totalCount)
	if totalCount == 0 {
		showing = "0 results"
	}

	return schedule.ListWorkScheduleResponse{
		TotalCount:    totalCount,
		Page:          filter.Page,
		Limit:         filter.Limit,
		TotalPages:    totalPages,
		Showing:       showing,
		WorkSchedules: workScheduleResponses,
	}, nil
}

// UpdateEmployeeScheduleAssignment implements schedule.ScheduleService.
func (s *scheduleServiceImpl) UpdateEmployeeScheduleAssignment(ctx context.Context, req schedule.UpdateEmployeeScheduleAssignmentRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	err = s.employeeScheduleAssignRepo.Update(ctx, req, companyID)
	if err != nil {
		if errors.Is(err, schedule.ErrEmployeeScheduleAssignmentNotFound) {
			return schedule.ErrEmployeeScheduleAssignmentNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return schedule.ErrOverlappingScheduleAssignment
		}
		return fmt.Errorf("failed to update employee schedule assignment: %w", err)
	}

	return nil
}

// UpdateWorkSchedule implements schedule.ScheduleService.
func (s *scheduleServiceImpl) UpdateWorkSchedule(ctx context.Context, req schedule.UpdateWorkScheduleRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	req.CompanyID = companyID

	postgresql.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		ws, err := s.workScheduleRepo.Update(txCtx, req)
		if err != nil {
			if errors.Is(err, schedule.ErrWorkScheduleNotFound) {
				return schedule.ErrWorkScheduleNotFound
			}

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return schedule.ErrWorkScheduleNameExists
			}

			return fmt.Errorf("failed to update work schedule: %w", err)
		}

		if req.Type != nil && ws.Type != schedule.WorkArrangementWFA && *req.Type == string(schedule.WorkArrangementWFA) {
			if err := s.workScheduleLocationRepo.BulkDeleteByWorkScheduleID(txCtx, ws.ID, companyID); err != nil {
				return fmt.Errorf("failed to delete work schedule locations: %w", err)
			}
		}

		return nil
	})

	return nil
}

// UpdateWorkScheduleLocation implements schedule.ScheduleService.
func (s *scheduleServiceImpl) UpdateWorkScheduleLocation(ctx context.Context, req schedule.UpdateWorkScheduleLocationRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	req.CompanyID = companyID

	err = s.workScheduleLocationRepo.Update(ctx, req)
	if err != nil {
		if errors.Is(err, schedule.ErrWorkScheduleLocationNotFound) {
			return schedule.ErrWorkScheduleLocationNotFound
		}
		return fmt.Errorf("failed to update work schedule location: %w", err)
	}

	return nil
}

// UpdateWorkScheduleTime implements schedule.ScheduleService.
func (s *scheduleServiceImpl) UpdateWorkScheduleTime(ctx context.Context, req schedule.UpdateWorkScheduleTimeRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	req.CompanyID = companyID

	// location type validation
	wsTimeData, err := s.workScheduleTimeRepo.GetByID(ctx, req.ID, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.ErrWorkScheduleTimeNotFound
		}
		return fmt.Errorf("failed to get work schedule time: %w", err)
	}

	wsData, err := s.workScheduleRepo.GetByID(ctx, wsTimeData.WorkScheduleID, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.ErrWorkScheduleNotFound
		}
		return fmt.Errorf("failed to get work schedule: %w", err)
	}

	// location type validation
	if wsData.Type == schedule.WorkArrangementWFO && req.LocationType != string(schedule.WorkArrangementWFO) {
		return schedule.ErrMismatchedLocationType
	}

	if wsData.Type == schedule.WorkArrangementWFA && req.LocationType != string(schedule.WorkArrangementWFA) {
		return schedule.ErrMismatchedLocationType
	}

	err = s.workScheduleTimeRepo.Update(ctx, req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return schedule.ErrWorkScheduleTimeExists
		}
		if errors.Is(err, schedule.ErrWorkScheduleTimeNotFound) {
			return schedule.ErrWorkScheduleTimeNotFound
		}
		return fmt.Errorf("failed to update work schedule time: %w", err)
	}

	return nil
}

func NewScheduleService(
	db *database.DB,
	workScheduleRepo schedule.WorkScheduleRepository,
	workScheduleTimeRepo schedule.WorkScheduleTimeRepository,
	workScheduleLocationRepo schedule.WorkScheduleLocationRepository,
	employeeScheduleAssignRepo schedule.EmployeeScheduleAssignmentRepository,
	employeeRepo employee.EmployeeRepository,
	notificationService notification.Service,
) schedule.ScheduleService {
	return &scheduleServiceImpl{
		db:                         db,
		workScheduleRepo:           workScheduleRepo,
		workScheduleTimeRepo:       workScheduleTimeRepo,
		workScheduleLocationRepo:   workScheduleLocationRepo,
		employeeScheduleAssignRepo: employeeScheduleAssignRepo,
		employeeRepo:               employeeRepo,
		notificationService:        notificationService,
	}
}

func (s *scheduleServiceImpl) mapWorkScheduleTimeToResponse(wst schedule.WorkScheduleTime) schedule.WorkScheduleTimeResponse {
	dayNames := []string{"", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

	var breakStart, breakEnd *string
	if wst.BreakStartTime != nil {
		bs := wst.BreakStartTime.Format("15:04")
		breakStart = &bs
	}
	if wst.BreakEndTime != nil {
		be := wst.BreakEndTime.Format("15:04")
		breakEnd = &be
	}

	return schedule.WorkScheduleTimeResponse{
		ID:                wst.ID,
		WorkScheduleID:    wst.WorkScheduleID,
		DayOfWeek:         wst.DayOfWeek,
		DayName:           dayNames[wst.DayOfWeek],
		ClockInTime:       wst.ClockInTime.Format("15:04"),
		BreakStartTime:    breakStart,
		BreakEndTime:      breakEnd,
		ClockOutTime:      wst.ClockOutTime.Format("15:04"),
		LocationType:      string(wst.LocationType),
		IsNextDayCheckout: wst.IsNextDayCheckout,
		CreatedAt:         wst.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         wst.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *scheduleServiceImpl) mapWorkScheduleLocationToResponse(wsl schedule.WorkScheduleLocation) schedule.WorkScheduleLocationResponse {
	return schedule.WorkScheduleLocationResponse{
		ID:             wsl.ID,
		WorkScheduleID: wsl.WorkScheduleID,
		LocationName:   wsl.LocationName,
		Latitude:       wsl.Latitude,
		Longitude:      wsl.Longitude,
		RadiusMeters:   wsl.RadiusMeters,
		CreatedAt:      wsl.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      wsl.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *scheduleServiceImpl) mapEmployeeScheduleAssignmentToResponse(esa schedule.EmployeeScheduleAssignment) schedule.EmployeeScheduleAssignmentResponse {
	return schedule.EmployeeScheduleAssignmentResponse{
		ID:             esa.ID,
		EmployeeID:     esa.EmployeeID,
		WorkScheduleID: esa.WorkScheduleID,
		StartDate:      esa.StartDate.Format("2006-01-02"),
		EndDate:        esa.EndDate.Format("2006-01-02"),
		CreatedAt:      esa.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      esa.UpdatedAt.Format(time.RFC3339),
	}
}

// GetEmployeeScheduleTimeline implements schedule.ScheduleService.
func (s *scheduleServiceImpl) GetEmployeeScheduleTimeline(ctx context.Context, employeeID string, filter schedule.EmployeeScheduleTimelineFilter) (schedule.EmployeeScheduleTimelineResponse, error) {
	// Extract company_id from JWT
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return schedule.EmployeeScheduleTimelineResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return schedule.EmployeeScheduleTimelineResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	// Validate filter
	if err := filter.Validate(); err != nil {
		return schedule.EmployeeScheduleTimelineResponse{}, err
	}

	// Get timeline data from repository
	items, total, employeeName, err := s.workScheduleRepo.GetEmployeeScheduleTimeline(ctx, employeeID, companyID, filter.Page, filter.Limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.EmployeeScheduleTimelineResponse{}, schedule.ErrEmployeeScheduleTimelineNotFound
		}
		return schedule.EmployeeScheduleTimelineResponse{}, fmt.Errorf("failed to get employee schedule timeline: %w", err)
	}

	// Calculate status and actions for each item
	today := time.Now()
	for i := range items {
		// Calculate status
		items[i].Status = s.calculateTimelineStatus(&items[i], today)

		// Calculate is_active_today
		items[i].IsActiveToday = s.isScheduleActiveToday(&items[i], today, items)

		// Calculate actions
		items[i].Actions = s.calculateTimelineActions(&items[i])
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))
	showing := s.calculateShowingText(filter.Page, filter.Limit, total)

	response := schedule.EmployeeScheduleTimelineResponse{
		EmployeeID:   employeeID,
		EmployeeName: employeeName,
		TotalCount:   total,
		Page:         filter.Page,
		Limit:        filter.Limit,
		TotalPages:   totalPages,
		Showing:      showing,
		Timeline:     items,
	}

	return response, nil
}

// calculateTimelineStatus determines the status of a timeline item
func (s *scheduleServiceImpl) calculateTimelineStatus(item *schedule.EmployeeScheduleTimelineItem, today time.Time) string {
	if item.Type == "default" {
		return "fallback"
	}

	// For override schedules
	if item.DateRange.Start != nil {
		startDate, _ := time.Parse("2006-01-02", *item.DateRange.Start)

		if item.DateRange.End != nil {
			endDate, _ := time.Parse("2006-01-02", *item.DateRange.End)

			if today.Before(startDate) {
				return "upcoming"
			} else if today.After(endDate) {
				return "past"
			} else {
				return "active"
			}
		}
	}

	return "active"
}

// isScheduleActiveToday checks if this schedule is actually being used today
func (s *scheduleServiceImpl) isScheduleActiveToday(item *schedule.EmployeeScheduleTimelineItem, today time.Time, allItems []schedule.EmployeeScheduleTimelineItem) bool {
	todayStr := today.Format("2006-01-02")

	// Check if this schedule covers today
	coversToday := false

	if item.Type == "default" {
		// Default always covers from start_date onwards
		if item.DateRange.Start != nil {
			if *item.DateRange.Start <= todayStr {
				coversToday = true
			}
		} else {
			coversToday = true // No start date means it's been active forever
		}
	} else {
		// Override schedule
		if item.DateRange.Start != nil && item.DateRange.End != nil {
			if *item.DateRange.Start <= todayStr && todayStr <= *item.DateRange.End {
				coversToday = true
			}
		}
	}

	if !coversToday {
		return false
	}

	// If this is an override that covers today, it's active
	if item.Type == "override" {
		return true
	}

	// If this is default, check if any override is active today
	for _, otherItem := range allItems {
		if otherItem.Type == "override" {
			if otherItem.DateRange.Start != nil && otherItem.DateRange.End != nil {
				if *otherItem.DateRange.Start <= todayStr && todayStr <= *otherItem.DateRange.End {
					return false // Override is active, so default is not active today
				}
			}
		}
	}

	return true // No active override, default is active
}

// calculateTimelineActions determines what actions can be performed on a timeline item
func (s *scheduleServiceImpl) calculateTimelineActions(item *schedule.EmployeeScheduleTimelineItem) schedule.ScheduleActions {
	if item.Type == "default" {
		return schedule.ScheduleActions{
			CanEdit:    false,
			CanDelete:  false,
			CanReplace: true,
		}
	}

	// Override schedules can be edited and deleted
	return schedule.ScheduleActions{
		CanEdit:   true,
		CanDelete: true,
	}
}

// calculateShowingText generates the "showing X-Y of Z results" text
func (s *scheduleServiceImpl) calculateShowingText(page, limit int, total int64) string {
	if total == 0 {
		return "0-0 of 0 results"
	}

	start := (page-1)*limit + 1
	end := start + limit - 1

	if end > int(total) {
		end = int(total)
	}

	return fmt.Sprintf("%d-%d of %d results", start, end, total)
}

// notifyEmployeeOnScheduleUpdated sends notification to employee when their schedule is assigned/updated
func (s *scheduleServiceImpl) notifyEmployeeOnScheduleUpdated(ctx context.Context, employeeID, workScheduleID, companyID, startDate string) {
	if s.notificationService == nil {
		return
	}

	// Get employee user ID
	emp, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil || emp.UserID == nil {
		return
	}

	// Get work schedule name
	ws, err := s.workScheduleRepo.GetByID(ctx, workScheduleID, companyID)
	if err != nil {
		return
	}

	_ = s.notificationService.QueueNotification(ctx, notification.CreateNotificationRequest{
		CompanyID:   companyID,
		RecipientID: *emp.UserID,
		SenderID:    nil,
		Type:        notification.TypeScheduleUpdated,
		Title:       "Schedule Updated",
		Message:     fmt.Sprintf("Your work schedule has been updated to '%s' starting from %s", ws.Name, startDate),
		Data: map[string]interface{}{
			"employee_id":      employeeID,
			"work_schedule_id": workScheduleID,
			"schedule_name":    ws.Name,
			"start_date":       startDate,
		},
	})
}
