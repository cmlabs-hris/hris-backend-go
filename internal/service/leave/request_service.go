package leave

import (
	"context"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type RequestService struct {
	db *database.DB
	leave.LeaveTypeRepository
	leave.LeaveQuotaRepository
	leave.LeaveRequestRepository
	employee.EmployeeRepository
}

func NewRequestService(db *database.DB, leaveTypeRepository leave.LeaveTypeRepository, leaveQuotaRepository leave.LeaveQuotaRepository, employeeRepository employee.EmployeeRepository) *RequestService {
	return &RequestService{
		db:                   db,
		LeaveTypeRepository:  leaveTypeRepository,
		LeaveQuotaRepository: leaveQuotaRepository,
		EmployeeRepository:   employeeRepository,
	}
}

func (r *RequestService) Approve(ctx context.Context, requestID string, approvedID string) (leave.LeaveRequest, error) {
	request, err := r.LeaveRequestRepository.GetByID(ctx, requestID)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to get leave request by ID: %w", err)
	}

	if request.Status != leave.LeaveRequestStatusWaitingApproval {
		return leave.LeaveRequest{}, leave.ErrLeaveAlreadyProcessed
	}

	request.Status = leave.LeaveRequestStatusApproved

	approvedAtTime := time.Now()
	request.Status = leave.LeaveRequestStatusApproved
	request.ApprovedAt = &approvedAtTime
	request.ApprovedBy = &approvedID

	updateStatus := string(request.Status)
	update := leave.UpdateLeaveRequestRequest{
		ID:         request.ID,
		Status:     &updateStatus,
		ApprovedBy: &approvedID,
		ApprovedAt: &approvedAtTime,
	}
	if err := r.LeaveRequestRepository.Update(ctx, update); err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to update leave request: %w", err)
	}

	return request, nil
}

func (r *RequestService) CreateRequest(ctx context.Context, req leave.CreateLeaveRequestRequest) (leave.LeaveRequest, error) {
	emp, err := r.EmployeeRepository.GetByUserID(ctx, req.EmployeeID)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to get employee by user ID: %w", err)
	}

	leaveType, err := r.LeaveTypeRepository.GetByID(ctx, req.LeaveTypeID)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to get leave type by ID: %w", err)
	}

	isEligiible, err := r.checkEligibility(ctx, emp, leaveType)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("eligibility check failed: %w", err)
	}
	if !isEligiible {
		return leave.LeaveRequest{}, leave.ErrNotEligible
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to parse start date: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to parse end date: %w", err)
	}

	if err := r.validateDates(ctx, leaveType, startDate, endDate); err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("date validation failed: %w", err)
	}

	hasOverlap, err := r.LeaveRequestRepository.CheckOverlapping(ctx, emp.ID, startDate, endDate)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to check overlapping leave requests: %w", err)
	}
	if hasOverlap {
		return leave.LeaveRequest{}, leave.ErrOverlappingLeave
	}

	workingDays, err := r.Calculate(ctx, emp.CompanyID, startDate, endDate, req.DurationType)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to calculate working days: %w", err)
	}

	request := leave.LeaveRequest{
		EmployeeID:    emp.ID,
		LeaveTypeID:   leaveType.ID,
		StartDate:     startDate,
		EndDate:       endDate,
		DurationType:  leave.LeaveDurationEnum(req.DurationType),
		TotalDays:     r.calculateTotalDays(startDate, endDate, req.DurationType),
		WorkingDays:   workingDays,
		Reason:        req.Reason,
		AttachmentURL: req.AttachmentURL,
		Status:        leave.LeaveRequestStatusWaitingApproval,
	}

	if startDate.Before(time.Now()) {
		request.IsBackdate = true
	}

	created, err := r.LeaveRequestRepository.Create(ctx, request)
	if err != nil {
		return leave.LeaveRequest{}, fmt.Errorf("failed to create leave request: %w", err)
	}

	return created, nil
}

func (r *RequestService) Reject(
	ctx context.Context, requestID, reason string, approvedID string) (leave.LeaveRequest, error) {
	request, err := r.LeaveRequestRepository.GetByID(ctx, requestID)
	if err != nil {
		return leave.LeaveRequest{}, err
	}

	if request.Status != leave.LeaveRequestStatusWaitingApproval {
		return leave.LeaveRequest{}, leave.ErrLeaveAlreadyProcessed
	}

	approvedAtTime := time.Now()
	request.Status = leave.LeaveRequestStatusRejected
	request.RejectionReason = &reason
	request.ApprovedBy = &approvedID
	request.ApprovedAt = &approvedAtTime

	updateStatus := string(request.Status)
	update := leave.UpdateLeaveRequestRequest{
		ID:              request.ID,
		Status:          &updateStatus,
		RejectionReason: request.RejectionReason,
		ApprovedBy:      request.ApprovedBy,
		ApprovedAt:      request.ApprovedAt,
	}
	err = r.LeaveRequestRepository.Update(ctx, update)
	if err != nil {
		return leave.LeaveRequest{}, err
	}

	return request, nil
}

func (r *RequestService) checkEligibility(ctx context.Context, emp employee.Employee, leaveType leave.LeaveType) (bool, error) {
	// Check if leave type is active
	if leaveType.IsActive != nil && !*leaveType.IsActive {
		return false, leave.ErrLeaveTypeInactive
	}

	// Check if employee has quota for this leave type (if applicable)
	if leaveType.HasQuota != nil && *leaveType.HasQuota {
		quota, err := r.LeaveQuotaRepository.GetByEmployeeTypeYear(ctx, emp.ID, leaveType.ID, time.Now().Year())
		if err != nil {
			return false, leave.ErrQuotaNotFound
		}

		// Check if employee has available quota
		if quota.AvailableQuota == nil || *quota.AvailableQuota <= 0 {
			return false, leave.ErrInsufficientQuota
		}
	}

	// Check eligibility based on quota calculation type
	switch leaveType.QuotaCalculationType {
	case "tenure":
		if !r.checkTenureEligibility(emp, &leaveType.QuotaRules) {
			return false, leave.ErrInsufficientTenure
		}
	case "position":
		if !r.checkPositionEligibility(emp, &leaveType.QuotaRules) {
			return false, leave.ErrPositionNotEligible
		}
	case "grade":
		if !r.checkGradeEligibility(emp, &leaveType.QuotaRules) {
			return false, leave.ErrGradeNotEligible
		}
	case "employment_type":
		if !r.checkEmploymentTypeEligibility(emp, &leaveType.QuotaRules) {
			return false, leave.ErrEmploymentTypeNotEligible
		}
	case "combined":
		if !r.checkCombinedEligibility(emp, &leaveType.QuotaRules) {
			return false, leave.ErrCombinedRequirementsNotMet
		}
	}

	return true, nil
}

func (r *RequestService) checkTenureEligibility(emp employee.Employee, rules *leave.QuotaRules) bool {
	if len(rules.Rules) == 0 {
		return true
	}

	tenureMonths := calculateTenureMonths(emp.HireDate)

	// Check if employee meets any of the tenure rules
	for _, rule := range rules.Rules {
		minMonths := 0
		if rule.MinMonths != nil {
			minMonths = *rule.MinMonths
		}

		maxMonths := 999999
		if rule.MaxMonths != nil {
			maxMonths = *rule.MaxMonths
		}

		if tenureMonths >= minMonths && tenureMonths < maxMonths {
			return true
		}
	}

	return false
}

func (r *RequestService) checkPositionEligibility(emp employee.Employee, rules *leave.QuotaRules) bool {
	if len(rules.Rules) == 0 {
		return true
	}

	// Check if employee's position matches any of the rules
	for _, rule := range rules.Rules {
		for _, positionID := range rule.PositionIDs {
			if positionID == emp.PositionID {
				return true
			}
		}
	}

	return false
}

func (r *RequestService) checkGradeEligibility(emp employee.Employee, rules *leave.QuotaRules) bool {
	if emp.GradeID == "" {
		return false
	}

	if len(rules.Rules) == 0 {
		return true
	}

	// Check if employee's grade matches any of the rules
	for _, rule := range rules.Rules {
		for _, gradeID := range rule.GradeIDs {
			if gradeID == emp.GradeID {
				return true
			}
		}
	}

	return false
}

func (r *RequestService) checkEmploymentTypeEligibility(emp employee.Employee, rules *leave.QuotaRules) bool {
	if len(rules.Rules) == 0 {
		return true
	}

	// Check if employee's employment type matches any of the rules
	for _, rule := range rules.Rules {
		if rule.EmploymentType == string(emp.EmploymentType) {
			return true
		}
	}

	return false
}

func (r *RequestService) checkCombinedEligibility(emp employee.Employee, rules *leave.QuotaRules) bool {
	if len(rules.Rules) == 0 {
		return true
	}

	tenureMonths := calculateTenureMonths(emp.HireDate)

	// Check if employee matches any of the combined rules
	for _, rule := range rules.Rules {
		if rule.Conditions == nil {
			continue
		}

		if r.matchesConditions(emp, tenureMonths, rule.Conditions) {
			return true
		}
	}

	return false
}

func (r *RequestService) matchesConditions(emp employee.Employee, tenureMonths int, conditions *leave.QuotaConditions) bool {
	// Check position IDs
	if len(conditions.PositionIDs) > 0 {
		matched := false
		for _, posID := range conditions.PositionIDs {
			if posID == emp.PositionID {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check grade IDs
	if len(conditions.GradeIDs) > 0 {
		if emp.GradeID == "" {
			return false
		}

		matched := false
		for _, gradeID := range conditions.GradeIDs {
			if gradeID == emp.GradeID {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check employment type
	if conditions.EmploymentType != "" {
		if conditions.EmploymentType != string(emp.EmploymentType) {
			return false
		}
	}

	// Check minimum tenure
	if conditions.MinTenureMonths != nil {
		if tenureMonths < *conditions.MinTenureMonths {
			return false
		}
	}

	// Check maximum tenure
	if conditions.MaxTenureMonths != nil {
		if tenureMonths >= *conditions.MaxTenureMonths {
			return false
		}
	}

	return true
}

func calculateTenureMonths(hireDate time.Time) int {
	now := time.Now()
	years := now.Year() - hireDate.Year()
	months := int(now.Month()) - int(hireDate.Month())
	totalMonths := years*12 + months

	if now.Day() < hireDate.Day() {
		totalMonths--
	}

	if totalMonths < 0 {
		totalMonths = 0
	}

	return totalMonths
}

// Validate date rules
func (r *RequestService) validateDates(
	ctx context.Context,
	leaveType leave.LeaveType,
	startDate, endDate time.Time,
) error {
	now := time.Now()

	// Check backdate
	if startDate.Before(now) {
		if leaveType.AllowBackdate == nil || !*leaveType.AllowBackdate {
			return leave.ErrBackdateNotAllowed
		}

		daysDiff := int(now.Sub(startDate).Hours() / 24)
		if leaveType.BackdateMaxDays != nil && daysDiff > *leaveType.BackdateMaxDays {
			return leave.ErrBackdateTooOld
		}
	}

	// Check notice period
	if leaveType.MinNoticeDays != nil {
		daysDiff := int(startDate.Sub(now).Hours() / 24)
		if daysDiff < *leaveType.MinNoticeDays {
			return leave.ErrInsufficientNotice
		}
	}

	// Check advance limit
	if leaveType.MaxAdvanceDays != nil {
		daysDiff := int(startDate.Sub(now).Hours() / 24)
		if daysDiff > *leaveType.MaxAdvanceDays {
			return leave.ErrTooFarAdvance
		}
	}

	// Check max days per request
	if leaveType.MaxDaysPerRequest != nil {
		totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
		if totalDays > *leaveType.MaxDaysPerRequest {
			return leave.ErrExceedsMaxDays
		}
	}

	return nil
}

// type WorkingDaysCalculator struct {
// 	holidayRepo leave.HolidayRepository
// }

// Calculate calculates working days excluding weekends and holidays
func (r *RequestService) Calculate(
	ctx context.Context,
	companyID string,
	startDate, endDate time.Time,
	durationType string,
) (float64, error) {
	// Get public holidays in date range
	// holidays, err := r.holidayRepo.GetByDateRange(ctx, companyID, startDate, endDate)
	// if err != nil {
	// 	return 0, err
	// }

	// Create holiday map for quick lookup
	// holidayMap := make(map[string]bool)
	// for _, h := range holidays {
	// 	holidayMap[h.Date.Format("2006-01-02")] = true
	// }

	var workingDays float64
	currentDate := startDate

	for !currentDate.After(endDate) {
		// Skip weekends (Saturday and Sunday)
		// if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
		// 	// Skip public holidays
		// 	// if !holidayMap[currentDate.Format("2006-01-02")] {
		// 	// 	// Handle half-day on first or last day
		// 	// 	if (currentDate.Equal(startDate) || currentDate.Equal(endDate)) &&
		// 	// 		(durationType == "half_day_morning" || durationType == "half_day_afternoon") {
		// 	// 		workingDays += 0.5
		// 	// 	} else {
		// 	// 		workingDays += 1.0
		// 	// 	}
		// 	// }
		// }

		if currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday {
			continue
		}
		// if holidayMap[currentDate.Format("2006-01-02")] {
		// 	continue
		// }
		if (currentDate.Equal(startDate) || currentDate.Equal(endDate)) &&
			(durationType == "half_day_morning" || durationType == "half_day_afternoon") {
			workingDays += 0.5
		} else {
			workingDays += 1.0
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return workingDays, nil
}

func (s *RequestService) calculateTotalDays(startDate, endDate time.Time, durationType string) float64 {
	days := float64(int(endDate.Sub(startDate).Hours()/24) + 1)

	if durationType == "half_day_morning" || durationType == "half_day_afternoon" {
		if days == 1 {
			return 0.5
		}
	}

	return days
}
