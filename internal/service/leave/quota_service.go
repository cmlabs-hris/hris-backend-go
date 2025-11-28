package leave

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
)

type QuotaService struct {
	db *database.DB
	leave.LeaveTypeRepository
	leave.LeaveQuotaRepository
	employee.EmployeeRepository
	calculator *QuotaCalculator
}

func NewQuotaService(db *database.DB, leaveTypeRepository leave.LeaveTypeRepository, leaveQuotaRepository leave.LeaveQuotaRepository, employeeRepository employee.EmployeeRepository, calculator *QuotaCalculator) *QuotaService {
	return &QuotaService{
		db:                   db,
		LeaveTypeRepository:  leaveTypeRepository,
		LeaveQuotaRepository: leaveQuotaRepository,
		EmployeeRepository:   employeeRepository,
		calculator:           calculator,
	}
}

// AssignLeaveQuotasForEmployee assigns leave quotas for a single employee based on all active leave types.
// This is used when creating a new employee (including owner) to automatically assign eligible leave quotas.
// It checks eligibility based on QuotaRules for each leave type.
func (q *QuotaService) AssignLeaveQuotasForEmployee(ctx context.Context, emp employee.Employee, year int) ([]leave.LeaveQuota, error) {
	// Get all active leave types for the company
	leaveTypes, err := q.LeaveTypeRepository.GetActiveByCompanyID(ctx, emp.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active leave types: %w", err)
	}

	assignedQuotas := make([]leave.LeaveQuota, 0)

	for _, leaveType := range leaveTypes {
		// Skip if leave type doesn't have quota
		if leaveType.HasQuota != nil && !*leaveType.HasQuota {
			slog.Debug("Skipping leave type without quota", "leave_type", leaveType.Name)
			continue
		}

		// Check if quota already exists for this employee and leave type
		existingQuota, err := q.LeaveQuotaRepository.GetByEmployeeTypeYear(ctx, emp.ID, leaveType.ID, year)
		if err == nil && existingQuota.ID != "" {
			slog.Debug("Quota already exists", "employee_id", emp.ID, "leave_type", leaveType.Name, "year", year)
			continue
		}

		// Calculate quota based on rules
		calculatedQuota, err := q.calculator.CalculateQuota(ctx, emp, leaveType)
		if err != nil {
			// Employee is not eligible for this leave type based on rules
			slog.Debug("Employee not eligible for leave type",
				"employee_id", emp.ID,
				"leave_type", leaveType.Name,
				"reason", err.Error(),
			)
			continue
		}

		// Skip if calculated quota is 0 or negative
		if calculatedQuota <= 0 {
			slog.Debug("Calculated quota is 0 or negative, skipping",
				"employee_id", emp.ID,
				"leave_type", leaveType.Name,
			)
			continue
		}

		// Determine opening balance and earned quota based on accrual method
		openingBalance := int(calculatedQuota)
		earnedQuota := 0

		if leaveType.AccrualMethod != nil && *leaveType.AccrualMethod == "monthly" {
			// For monthly accrual, calculate pro-rated quota
			accruedQuota := q.calculator.CalculateAccruedQuota(emp.HireDate, calculatedQuota, time.Now())
			openingBalance = 0
			earnedQuota = int(accruedQuota)
		}

		// Create the quota
		zeroFloat := 0.0
		zeroInt := 0
		newQuota := leave.LeaveQuota{
			EmployeeID:      emp.ID,
			LeaveTypeID:     leaveType.ID,
			Year:            year,
			OpeningBalance:  &openingBalance,
			EarnedQuota:     &earnedQuota,
			RolloverQuota:   &zeroInt,
			AdjustmentQuota: &zeroInt,
			UsedQuota:       &zeroFloat,
			PendingQuota:    &zeroFloat,
		}

		createdQuota, err := q.LeaveQuotaRepository.Create(ctx, newQuota)
		if err != nil {
			slog.Warn("Failed to create quota",
				"employee_id", emp.ID,
				"leave_type", leaveType.Name,
				"error", err,
			)
			continue
		}

		assignedQuotas = append(assignedQuotas, createdQuota)
		slog.Info("Assigned leave quota",
			"employee_id", emp.ID,
			"leave_type", leaveType.Name,
			"opening_balance", openingBalance,
			"earned_quota", earnedQuota,
			"year", year,
		)
	}

	return assignedQuotas, nil
}

// AssignLeaveQuotasForEmployeeByID is a convenience method that fetches the employee first
func (q *QuotaService) AssignLeaveQuotasForEmployeeByID(ctx context.Context, employeeID string, year int) ([]leave.LeaveQuota, error) {
	emp, err := q.EmployeeRepository.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	return q.AssignLeaveQuotasForEmployee(ctx, emp, year)
}

func (q *QuotaService) AllocateTypeQuota(ctx context.Context, leaveType leave.LeaveType, companyID string, year int) error {
	employees, err := q.EmployeeRepository.GetActiveByCompanyID(ctx, companyID)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}

	for _, employee := range employees {
		exists, err := q.LeaveQuotaRepository.GetByEmployeeTypeYear(ctx, employee.ID, leaveType.ID, year)
		if err != nil {
			if err != pgx.ErrNoRows {
				return fmt.Errorf("failed to get leave quota: %w", err)
			}

			if exists.ID != "" {
				// quota of this leave type already exists
				continue
			}

			calculatedQuota, err := q.calculator.CalculateQuota(ctx, employee, leaveType)
			if err != nil {
				// not eligible
				continue
			}

			openingBalance := int(calculatedQuota)
			earnedQuota := 0

			if leaveType.AccrualMethod != nil && *leaveType.AccrualMethod == "monthly" {
				accruedQuota := q.calculator.CalculateAccruedQuota(employee.HireDate, calculatedQuota, time.Now())
				openingBalance = 0
				earnedQuota = int(accruedQuota)
			}

			newQuota := leave.LeaveQuota{
				EmployeeID:      employee.ID,
				LeaveTypeID:     leaveType.ID,
				Year:            year,
				OpeningBalance:  &openingBalance,
				EarnedQuota:     &earnedQuota,
				RolloverQuota:   nil,
				AdjustmentQuota: nil,
				UsedQuota:       nil,
				PendingQuota:    nil,
			}

			_, err = q.LeaveQuotaRepository.Create(ctx, newQuota)
			if err != nil {
				return fmt.Errorf("failed to create quota for %s: %w", leaveType.Name, err)
			}
		}
	}

	return nil
}

// AdjustQuota allows manual adjustment by HR
func (q *QuotaService) AdjustQuota(
	ctx context.Context,
	employeeID, leaveTypeID string,
	year int,
	adjustment int,
	reason string,
) error {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	adjustedBy, ok := claims["user_id"].(string)
	if !ok || adjustedBy == "" {
		return fmt.Errorf("user_id claim is missing or invalid")
	}

	quota, err := q.LeaveQuotaRepository.GetByEmployeeTypeYear(
		ctx,
		employeeID,
		leaveTypeID,
		year,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return leave.ErrQuotaNotFound
		}
		return fmt.Errorf("failed to fetch quota: %w", err)
	}

	oldAdjustment := 0
	if quota.AdjustmentQuota != nil {
		oldAdjustment = *quota.AdjustmentQuota
	}

	newAdjustment := oldAdjustment + adjustment
	quota.AdjustmentQuota = &newAdjustment

	newAvailable := int(*quota.OpeningBalance) + int(*quota.EarnedQuota) +
		int(*quota.RolloverQuota) + int(*quota.AdjustmentQuota) -
		int(*quota.UsedQuota) - int(*quota.PendingQuota)

	if newAvailable < 0 {
		return leave.ErrNegativeQuota
	}

	updateRequest := leave.UpdateLeaveQuotaRequest{
		ID:              quota.ID,
		AdjustmentQuota: quota.AdjustmentQuota,
	}

	err = q.LeaveQuotaRepository.Update(ctx, updateRequest)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	// Log the adjustment
	fmt.Printf("Quota adjusted for employee %s, leave type %s: old adjustment %d, new adjustment %d, reason: %s, adjusted by: %s\n",
		employeeID, leaveTypeID, oldAdjustment, newAdjustment, reason, adjustedBy)

	return nil
}

// RecalculateQuota recalculates quota based on current rules (for existing quotas)
func (s *QuotaService) RecalculateQuota(
	ctx context.Context,
	employeeID, leaveTypeID string,
	year int,
) error {
	// Get employee and leave type
	emp, err := s.EmployeeRepository.GetByID(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}

	leaveType, err := s.LeaveTypeRepository.GetByID(ctx, leaveTypeID)
	if err != nil {
		return fmt.Errorf("failed to get leave type: %w", err)
	}

	// Get existing quota
	quota, err := s.LeaveQuotaRepository.GetByEmployeeTypeYear(ctx, employeeID, leaveTypeID, year)
	if err != nil {
		return fmt.Errorf("failed to get leave quota: %w", err)
	}
	// Calculate new quota
	oldOpening := 0
	if quota.OpeningBalance != nil {
		oldOpening = *quota.OpeningBalance
	}

	calculatedQuota, err := s.calculator.CalculateQuota(ctx, emp, leaveType)
	if err != nil {
		// not eligible
		return fmt.Errorf("failed to calculate new quota: %w", err)
	}

	openingBalance := int(calculatedQuota)
	earnedQuota := 0

	if leaveType.AccrualMethod != nil && *leaveType.AccrualMethod == "monthly" {
		accruedQuota := s.calculator.CalculateAccruedQuota(emp.HireDate, calculatedQuota, time.Now())
		openingBalance = 0
		earnedQuota = int(accruedQuota)
	}

	quota.OpeningBalance = &openingBalance
	quota.EarnedQuota = &earnedQuota

	updateRequest := leave.UpdateLeaveQuotaRequest{
		ID:             quota.ID,
		OpeningBalance: quota.OpeningBalance,
		EarnedQuota:    quota.EarnedQuota,
	}

	err = s.LeaveQuotaRepository.Update(ctx, updateRequest)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	// Log the recalculation
	fmt.Printf("Quota recalculated for employee %s, leave type %s: old opening balance %d, new opening balance %d\n",
		employeeID, leaveType.Name, oldOpening, openingBalance)

	return nil
}

// MovePendingToUsed moves pending leave days to used leave days upon approval
func (q *QuotaService) MovePendingToUsed(
	ctx context.Context,
	employeeID, leaveTypeID string,
	days float64,
) error {
	quota, err := q.LeaveQuotaRepository.GetByEmployeeTypeYear(
		ctx,
		employeeID,
		leaveTypeID,
		time.Now().Year(),
	)
	if err != nil {
		return fmt.Errorf("failed to fetch quota: %w", err)
	}

	err = q.LeaveQuotaRepository.MovePendingToUsed(ctx, quota.ID, days)
	if err != nil {
		return fmt.Errorf("failed to move pending to used: %w", err)
	}

	// Log the operation
	fmt.Printf("Moved %.2f days from pending to used for employee %s, leave type %s\n",
		days, employeeID, leaveTypeID)

	return nil
}

// ReleaseQuota releases pending quota (on rejection/cancellation)
func (q *QuotaService) ReleaseQuota(
	ctx context.Context,
	employeeID, leaveTypeID string,
	days float64,
) error {
	quota, err := q.LeaveQuotaRepository.GetByEmployeeTypeYear(
		ctx,
		employeeID,
		leaveTypeID,
		time.Now().Year(),
	)
	if err != nil {
		return fmt.Errorf("failed to fetch quota: %w", err)
	}

	err = q.LeaveQuotaRepository.RemovePendingQuota(ctx, quota.ID, days)
	if err != nil {
		return fmt.Errorf("failed to release pending quota: %w", err)
	}

	// Log the operation
	fmt.Printf("Released %.2f days from pending quota for employee %s, leave type %s\n",
		days, employeeID, leaveTypeID)

	return nil
}

// ReserveQuota reserves quota for a pending request
func (q *QuotaService) ReserveQuota(
	ctx context.Context,
	employeeID, leaveTypeID string,
	days float64,
) error {
	quota, err := q.LeaveQuotaRepository.GetByEmployeeTypeYear(
		ctx,
		employeeID,
		leaveTypeID,
		time.Now().Year(),
	)
	if err != nil {
		return fmt.Errorf("failed to fetch quota: %w", err)
	}

	// Check available quota
	available := float64(*quota.OpeningBalance) + float64(*quota.EarnedQuota) + float64(*quota.RolloverQuota) + float64(*quota.AdjustmentQuota) - *quota.UsedQuota - *quota.PendingQuota
	if available < days {
		return leave.ErrInsufficientQuota
	}

	// Add to pending
	err = q.LeaveQuotaRepository.AddPendingQuota(ctx, quota.ID, days)
	if err != nil {
		return fmt.Errorf("failed to reserve quota: %w", err)
	}

	// Log the operation
	fmt.Printf("Reserved %.2f days from available quota for employee %s, leave type %s\n",
		days, employeeID, leaveTypeID)

	return nil
}
