package payroll

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/payroll"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type PayrollServiceImpl struct {
	db           *database.DB
	payrollRepo  payroll.PayrollRepository
	employeeRepo employee.EmployeeRepository
}

func NewPayrollService(
	db *database.DB,
	payrollRepo payroll.PayrollRepository,
	employeeRepo employee.EmployeeRepository,
) payroll.PayrollService {
	return &PayrollServiceImpl{
		db:           db,
		payrollRepo:  payrollRepo,
		employeeRepo: employeeRepo,
	}
}

// Helper to get company_id and user_id from JWT context
func getClaimsFromContext(ctx context.Context) (companyID, userID string, err error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return "", "", fmt.Errorf("company_id claim is missing or invalid")
	}

	userID, _ = claims["user_id"].(string)

	return companyID, userID, nil
}

// ========== SETTINGS ==========

func (s *PayrollServiceImpl) GetSettings(ctx context.Context) (payroll.PayrollSettingsResponse, error) {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.PayrollSettingsResponse{}, err
	}

	settings, err := s.payrollRepo.GetSettings(ctx, companyID)
	if err != nil {
		if errors.Is(err, payroll.ErrPayrollSettingsNotFound) {
			// Return default settings
			return payroll.PayrollSettingsResponse{
				CompanyID:                    companyID,
				LateDeductionEnabled:         true,
				LateDeductionPerMinute:       decimal.Zero,
				OvertimeEnabled:              true,
				OvertimePayPerMinute:         decimal.Zero,
				EarlyLeaveDeductionEnabled:   false,
				EarlyLeaveDeductionPerMinute: decimal.Zero,
			}, nil
		}
		return payroll.PayrollSettingsResponse{}, err
	}

	return payroll.PayrollSettingsResponse{
		ID:                           settings.ID,
		CompanyID:                    settings.CompanyID,
		LateDeductionEnabled:         settings.LateDeductionEnabled,
		LateDeductionPerMinute:       settings.LateDeductionPerMinute,
		OvertimeEnabled:              settings.OvertimeEnabled,
		OvertimePayPerMinute:         settings.OvertimePayPerMinute,
		EarlyLeaveDeductionEnabled:   settings.EarlyLeaveDeductionEnabled,
		EarlyLeaveDeductionPerMinute: settings.EarlyLeaveDeductionPerMinute,
	}, nil
}

func (s *PayrollServiceImpl) UpdateSettings(ctx context.Context, req payroll.UpdatePayrollSettingsRequest) (payroll.PayrollSettingsResponse, error) {
	if err := req.Validate(); err != nil {
		return payroll.PayrollSettingsResponse{}, err
	}

	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.PayrollSettingsResponse{}, err
	}

	// Get current settings or use defaults
	current, err := s.payrollRepo.GetSettings(ctx, companyID)
	if err != nil && !errors.Is(err, payroll.ErrPayrollSettingsNotFound) {
		return payroll.PayrollSettingsResponse{}, err
	}

	// If not found, initialize with defaults
	if errors.Is(err, payroll.ErrPayrollSettingsNotFound) {
		current = payroll.PayrollSettings{
			CompanyID:                    companyID,
			LateDeductionEnabled:         true,
			LateDeductionPerMinute:       decimal.Zero,
			OvertimeEnabled:              true,
			OvertimePayPerMinute:         decimal.Zero,
			EarlyLeaveDeductionEnabled:   false,
			EarlyLeaveDeductionPerMinute: decimal.Zero,
		}
	}

	// Apply updates
	if req.LateDeductionEnabled != nil {
		current.LateDeductionEnabled = *req.LateDeductionEnabled
	}
	if req.LateDeductionPerMinute != nil {
		current.LateDeductionPerMinute = *req.LateDeductionPerMinute
	}
	if req.OvertimeEnabled != nil {
		current.OvertimeEnabled = *req.OvertimeEnabled
	}
	if req.OvertimePayPerMinute != nil {
		current.OvertimePayPerMinute = *req.OvertimePayPerMinute
	}
	if req.EarlyLeaveDeductionEnabled != nil {
		current.EarlyLeaveDeductionEnabled = *req.EarlyLeaveDeductionEnabled
	}
	if req.EarlyLeaveDeductionPerMinute != nil {
		current.EarlyLeaveDeductionPerMinute = *req.EarlyLeaveDeductionPerMinute
	}

	updated, err := s.payrollRepo.UpsertSettings(ctx, current)
	if err != nil {
		return payroll.PayrollSettingsResponse{}, err
	}

	return payroll.PayrollSettingsResponse{
		ID:                           updated.ID,
		CompanyID:                    updated.CompanyID,
		LateDeductionEnabled:         updated.LateDeductionEnabled,
		LateDeductionPerMinute:       updated.LateDeductionPerMinute,
		OvertimeEnabled:              updated.OvertimeEnabled,
		OvertimePayPerMinute:         updated.OvertimePayPerMinute,
		EarlyLeaveDeductionEnabled:   updated.EarlyLeaveDeductionEnabled,
		EarlyLeaveDeductionPerMinute: updated.EarlyLeaveDeductionPerMinute,
	}, nil
}

// ========== COMPONENTS ==========

func (s *PayrollServiceImpl) CreateComponent(ctx context.Context, req payroll.CreatePayrollComponentRequest) (payroll.PayrollComponentResponse, error) {
	if err := req.Validate(); err != nil {
		return payroll.PayrollComponentResponse{}, err
	}

	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.PayrollComponentResponse{}, err
	}

	isTaxable := false
	if req.IsTaxable != nil {
		isTaxable = *req.IsTaxable
	}

	component := payroll.PayrollComponent{
		CompanyID:   companyID,
		Name:        req.Name,
		Type:        payroll.ComponentType(req.Type),
		Description: req.Description,
		IsTaxable:   isTaxable,
		IsActive:    true,
	}

	created, err := s.payrollRepo.CreateComponent(ctx, component)
	if err != nil {
		return payroll.PayrollComponentResponse{}, err
	}

	return payroll.PayrollComponentResponse{
		ID:          created.ID,
		CompanyID:   created.CompanyID,
		Name:        created.Name,
		Type:        string(created.Type),
		Description: created.Description,
		IsTaxable:   created.IsTaxable,
		IsActive:    created.IsActive,
	}, nil
}

func (s *PayrollServiceImpl) GetComponent(ctx context.Context, id string) (payroll.PayrollComponentResponse, error) {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.PayrollComponentResponse{}, err
	}

	component, err := s.payrollRepo.GetComponentByID(ctx, id, companyID)
	if err != nil {
		return payroll.PayrollComponentResponse{}, err
	}

	return payroll.PayrollComponentResponse{
		ID:          component.ID,
		CompanyID:   component.CompanyID,
		Name:        component.Name,
		Type:        string(component.Type),
		Description: component.Description,
		IsTaxable:   component.IsTaxable,
		IsActive:    component.IsActive,
	}, nil
}

func (s *PayrollServiceImpl) ListComponents(ctx context.Context, activeOnly bool) ([]payroll.PayrollComponentResponse, error) {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	components, err := s.payrollRepo.GetComponentsByCompanyID(ctx, companyID, activeOnly)
	if err != nil {
		return nil, err
	}

	var result []payroll.PayrollComponentResponse
	for _, c := range components {
		result = append(result, payroll.PayrollComponentResponse{
			ID:          c.ID,
			CompanyID:   c.CompanyID,
			Name:        c.Name,
			Type:        string(c.Type),
			Description: c.Description,
			IsTaxable:   c.IsTaxable,
			IsActive:    c.IsActive,
		})
	}

	return result, nil
}

func (s *PayrollServiceImpl) UpdateComponent(ctx context.Context, req payroll.UpdatePayrollComponentRequest) error {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return err
	}

	return s.payrollRepo.UpdateComponent(ctx, companyID, req)
}

func (s *PayrollServiceImpl) DeleteComponent(ctx context.Context, id string) error {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return err
	}

	return s.payrollRepo.DeleteComponent(ctx, id, companyID)
}

// ========== EMPLOYEE COMPONENTS ==========

func (s *PayrollServiceImpl) AssignComponentToEmployee(ctx context.Context, req payroll.AssignComponentRequest) (payroll.EmployeeComponentResponse, error) {
	if err := req.Validate(); err != nil {
		return payroll.EmployeeComponentResponse{}, err
	}

	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.EmployeeComponentResponse{}, err
	}

	effectiveDate := time.Now()
	if req.EffectiveDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.EffectiveDate)
		if err == nil {
			effectiveDate = parsed
		}
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err == nil {
			endDate = &parsed
		}
	}

	assignment := payroll.EmployeePayrollComponent{
		EmployeeID:         req.EmployeeID,
		PayrollComponentID: req.PayrollComponentID,
		Amount:             req.Amount,
		EffectiveDate:      effectiveDate,
		EndDate:            endDate,
	}

	created, err := s.payrollRepo.AssignComponentToEmployee(ctx, assignment, companyID)
	if err != nil {
		return payroll.EmployeeComponentResponse{}, err
	}

	// Get component details
	comp, _ := s.payrollRepo.GetComponentByID(ctx, created.PayrollComponentID, companyID)

	var endDateStr *string
	if created.EndDate != nil {
		str := created.EndDate.Format("2006-01-02")
		endDateStr = &str
	}

	return payroll.EmployeeComponentResponse{
		ID:                 created.ID,
		EmployeeID:         created.EmployeeID,
		PayrollComponentID: created.PayrollComponentID,
		ComponentName:      comp.Name,
		ComponentType:      string(comp.Type),
		Amount:             created.Amount,
		EffectiveDate:      created.EffectiveDate.Format("2006-01-02"),
		EndDate:            endDateStr,
	}, nil
}

func (s *PayrollServiceImpl) GetEmployeeComponents(ctx context.Context, employeeID string) ([]payroll.EmployeeComponentResponse, error) {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	assignments, err := s.payrollRepo.GetEmployeeComponents(ctx, employeeID, companyID, false)
	if err != nil {
		return nil, err
	}

	var result []payroll.EmployeeComponentResponse
	for _, a := range assignments {
		var endDateStr *string
		if a.EndDate != nil {
			str := a.EndDate.Format("2006-01-02")
			endDateStr = &str
		}

		componentName := ""
		componentType := ""
		if a.ComponentName != nil {
			componentName = *a.ComponentName
		}
		if a.ComponentType != nil {
			componentType = string(*a.ComponentType)
		}

		result = append(result, payroll.EmployeeComponentResponse{
			ID:                 a.ID,
			EmployeeID:         a.EmployeeID,
			PayrollComponentID: a.PayrollComponentID,
			ComponentName:      componentName,
			ComponentType:      componentType,
			Amount:             a.Amount,
			EffectiveDate:      a.EffectiveDate.Format("2006-01-02"),
			EndDate:            endDateStr,
		})
	}

	return result, nil
}

func (s *PayrollServiceImpl) UpdateEmployeeComponent(ctx context.Context, req payroll.UpdateEmployeeComponentRequest) error {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return err
	}

	return s.payrollRepo.UpdateEmployeeComponent(ctx, companyID, req)
}

func (s *PayrollServiceImpl) RemoveEmployeeComponent(ctx context.Context, id string) error {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return err
	}

	return s.payrollRepo.RemoveEmployeeComponent(ctx, id, companyID)
}

// ========== PAYROLL GENERATION ==========

func (s *PayrollServiceImpl) GeneratePayroll(ctx context.Context, req payroll.GeneratePayrollRequest) ([]payroll.PayrollRecordResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get payroll settings
	settings, err := s.payrollRepo.GetSettings(ctx, companyID)
	if err != nil && !errors.Is(err, payroll.ErrPayrollSettingsNotFound) {
		return nil, err
	}
	// If not found, use defaults
	if errors.Is(err, payroll.ErrPayrollSettingsNotFound) {
		settings = payroll.PayrollSettings{
			LateDeductionEnabled:         true,
			LateDeductionPerMinute:       decimal.Zero,
			OvertimeEnabled:              true,
			OvertimePayPerMinute:         decimal.Zero,
			EarlyLeaveDeductionEnabled:   false,
			EarlyLeaveDeductionPerMinute: decimal.Zero,
		}
	}

	// Get employees
	var employees []employee.Employee
	if len(req.EmployeeIDs) > 0 {
		// TODO: Get employees by IDs - for now, get all and filter
		allEmployees, err := s.employeeRepo.GetActiveByCompanyID(ctx, companyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get employees: %w", err)
		}
		employeeIDSet := make(map[string]bool)
		for _, id := range req.EmployeeIDs {
			employeeIDSet[id] = true
		}
		for _, emp := range allEmployees {
			if employeeIDSet[emp.ID] {
				employees = append(employees, emp)
			}
		}
	} else {
		employees, err = s.employeeRepo.GetActiveByCompanyID(ctx, companyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get employees: %w", err)
		}
	}

	// Get attendance summary
	var employeeIDs []string
	for _, emp := range employees {
		employeeIDs = append(employeeIDs, emp.ID)
	}
	attendanceSummaries, err := s.payrollRepo.GetAttendanceSummary(ctx, companyID, req.PeriodMonth, req.PeriodYear, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance summary: %w", err)
	}
	attendanceMap := make(map[string]payroll.AttendanceSummary)
	for _, a := range attendanceSummaries {
		attendanceMap[a.EmployeeID] = a
	}

	// Generate payroll for each employee
	var records []payroll.PayrollRecord
	for _, emp := range employees {
		if emp.BaseSalary == nil || emp.BaseSalary.IsZero() {
			continue // Skip employees without base salary
		}

		// Check if record already exists
		_, err := s.payrollRepo.GetPayrollRecordByEmployeePeriod(ctx, emp.ID, req.PeriodMonth, req.PeriodYear, companyID)
		if err == nil {
			continue // Skip if already exists
		}
		if !errors.Is(err, payroll.ErrPayrollRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing payroll record: %w", err)
		}

		// Get employee components
		components, _ := s.payrollRepo.GetEmployeeComponents(ctx, emp.ID, companyID, true)

		totalAllowances := decimal.Zero
		totalDeductions := decimal.Zero
		allowancesDetail := make(map[string]decimal.Decimal)
		deductionsDetail := make(map[string]decimal.Decimal)

		for _, comp := range components {
			if comp.ComponentType != nil {
				if *comp.ComponentType == payroll.ComponentTypeAllowance {
					totalAllowances = totalAllowances.Add(comp.Amount)
					if comp.ComponentName != nil {
						allowancesDetail[*comp.ComponentName] = comp.Amount
					}
				} else {
					totalDeductions = totalDeductions.Add(comp.Amount)
					if comp.ComponentName != nil {
						deductionsDetail[*comp.ComponentName] = comp.Amount
					}
				}
			}
		}

		// Get attendance data
		att := attendanceMap[emp.ID]

		// Calculate late/overtime deductions using decimal
		lateDeduction := decimal.Zero
		earlyLeaveDeduction := decimal.Zero
		overtimeAmount := decimal.Zero

		if settings.LateDeductionEnabled {
			lateDeduction = decimal.NewFromInt(int64(att.TotalLateMinutes)).Mul(settings.LateDeductionPerMinute)
		}
		if settings.EarlyLeaveDeductionEnabled {
			earlyLeaveDeduction = decimal.NewFromInt(int64(att.TotalEarlyLeaveMinutes)).Mul(settings.EarlyLeaveDeductionPerMinute)
		}
		if settings.OvertimeEnabled {
			overtimeAmount = decimal.NewFromInt(int64(att.TotalOvertimeMinutes)).Mul(settings.OvertimePayPerMinute)
		}

		// Calculate final salary using decimal arithmetic
		grossSalary := emp.BaseSalary.Add(totalAllowances).Add(overtimeAmount)
		netSalary := grossSalary.Sub(totalDeductions).Sub(lateDeduction).Sub(earlyLeaveDeduction)

		record := payroll.PayrollRecord{
			EmployeeID:                emp.ID,
			CompanyID:                 companyID,
			PeriodMonth:               req.PeriodMonth,
			PeriodYear:                req.PeriodYear,
			BaseSalary:                *emp.BaseSalary,
			TotalAllowances:           totalAllowances,
			TotalDeductions:           totalDeductions,
			AllowancesDetail:          allowancesDetail,
			DeductionsDetail:          deductionsDetail,
			TotalWorkDays:             att.TotalWorkDays,
			TotalLateMinutes:          att.TotalLateMinutes,
			LateDeductionAmount:       lateDeduction,
			TotalEarlyLeaveMinutes:    att.TotalEarlyLeaveMinutes,
			EarlyLeaveDeductionAmount: earlyLeaveDeduction,
			TotalOvertimeMinutes:      att.TotalOvertimeMinutes,
			OvertimeAmount:            overtimeAmount,
			GrossSalary:               grossSalary,
			NetSalary:                 netSalary,
			Status:                    payroll.PayrollStatusDraft,
		}

		created, err := s.payrollRepo.CreatePayrollRecord(ctx, record)
		if err != nil {
			if errors.Is(err, payroll.ErrPayrollRecordAlreadyExists) {
				continue
			}
			return nil, fmt.Errorf("failed to create payroll record for employee %s: %w", emp.ID, err)
		}
		records = append(records, created)
	}

	return mapToRecordResponses(records), nil
}

func (s *PayrollServiceImpl) GetPayrollRecord(ctx context.Context, id string) (payroll.PayrollRecordResponse, error) {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.PayrollRecordResponse{}, err
	}

	record, err := s.payrollRepo.GetPayrollRecordByID(ctx, id, companyID)
	if err != nil {
		return payroll.PayrollRecordResponse{}, err
	}

	return mapToRecordResponse(record), nil
}

func (s *PayrollServiceImpl) ListPayrollRecords(ctx context.Context, filter payroll.PayrollFilter) (payroll.ListPayrollRecordResponse, error) {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.ListPayrollRecordResponse{}, err
	}

	records, totalCount, err := s.payrollRepo.ListPayrollRecords(ctx, companyID, filter)
	if err != nil {
		return payroll.ListPayrollRecordResponse{}, err
	}

	return payroll.ListPayrollRecordResponse{
		Data:       mapToRecordResponses(records),
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

func (s *PayrollServiceImpl) UpdatePayrollRecord(ctx context.Context, req payroll.UpdatePayrollRecordRequest) (payroll.PayrollRecordResponse, error) {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.PayrollRecordResponse{}, err
	}

	if err := s.payrollRepo.UpdatePayrollRecord(ctx, companyID, req); err != nil {
		return payroll.PayrollRecordResponse{}, err
	}

	return s.GetPayrollRecord(ctx, req.ID)
}

func (s *PayrollServiceImpl) FinalizePayroll(ctx context.Context, req payroll.FinalizePayrollRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	companyID, userID, err := getClaimsFromContext(ctx)
	if err != nil {
		return err
	}

	return s.payrollRepo.FinalizePayrollRecords(ctx, req.RecordIDs, userID, companyID)
}

func (s *PayrollServiceImpl) DeletePayrollRecord(ctx context.Context, id string) error {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return err
	}

	return s.payrollRepo.DeletePayrollRecord(ctx, id, companyID)
}

// ========== SUMMARY ==========

func (s *PayrollServiceImpl) GetPayrollSummary(ctx context.Context, month, year int) (payroll.PayrollSummaryResponse, error) {
	companyID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return payroll.PayrollSummaryResponse{}, err
	}

	return s.payrollRepo.GetPayrollSummary(ctx, companyID, month, year)
}

// ========== HELPERS ==========

func mapToRecordResponse(r payroll.PayrollRecord) payroll.PayrollRecordResponse {
	var paidAtStr *string
	if r.PaidAt != nil {
		str := r.PaidAt.Format(time.RFC3339)
		paidAtStr = &str
	}

	employeeName := ""
	employeeCode := ""
	if r.EmployeeName != nil {
		employeeName = *r.EmployeeName
	}
	if r.EmployeeCode != nil {
		employeeCode = *r.EmployeeCode
	}

	return payroll.PayrollRecordResponse{
		ID:                        r.ID,
		EmployeeID:                r.EmployeeID,
		EmployeeName:              employeeName,
		EmployeeCode:              employeeCode,
		PositionName:              r.PositionName,
		BranchName:                r.BranchName,
		PeriodMonth:               r.PeriodMonth,
		PeriodYear:                r.PeriodYear,
		BaseSalary:                r.BaseSalary,
		TotalAllowances:           r.TotalAllowances,
		TotalDeductions:           r.TotalDeductions,
		AllowancesDetail:          r.AllowancesDetail,
		DeductionsDetail:          r.DeductionsDetail,
		TotalWorkDays:             r.TotalWorkDays,
		TotalLateMinutes:          r.TotalLateMinutes,
		LateDeductionAmount:       r.LateDeductionAmount,
		TotalEarlyLeaveMinutes:    r.TotalEarlyLeaveMinutes,
		EarlyLeaveDeductionAmount: r.EarlyLeaveDeductionAmount,
		TotalOvertimeMinutes:      r.TotalOvertimeMinutes,
		OvertimeAmount:            r.OvertimeAmount,
		GrossSalary:               r.GrossSalary,
		NetSalary:                 r.NetSalary,
		Status:                    string(r.Status),
		PaidAt:                    paidAtStr,
		Notes:                     r.Notes,
	}
}

func mapToRecordResponses(records []payroll.PayrollRecord) []payroll.PayrollRecordResponse {
	result := make([]payroll.PayrollRecordResponse, 0, len(records))
	for _, r := range records {
		result = append(result, mapToRecordResponse(r))
	}
	return result
}

// Ensure pgx.ErrNoRows is handled properly
var _ = pgx.ErrNoRows
