package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/payroll"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type payrollRepository struct {
	db *database.DB
}

func NewPayrollRepository(db *database.DB) payroll.PayrollRepository {
	return &payrollRepository{db: db}
}

// ========== SETTINGS ==========

func (r *payrollRepository) GetSettings(ctx context.Context, companyID string) (payroll.PayrollSettings, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, late_deduction_enabled, late_deduction_per_minute,
			   overtime_enabled, overtime_pay_per_minute,
			   early_leave_deduction_enabled, early_leave_deduction_per_minute,
			   created_at, updated_at
		FROM payroll_settings
		WHERE company_id = $1
	`

	var s payroll.PayrollSettings
	err := q.QueryRow(ctx, query, companyID).Scan(
		&s.ID, &s.CompanyID, &s.LateDeductionEnabled, &s.LateDeductionPerMinute,
		&s.OvertimeEnabled, &s.OvertimePayPerMinute,
		&s.EarlyLeaveDeductionEnabled, &s.EarlyLeaveDeductionPerMinute,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.PayrollSettings{}, payroll.ErrPayrollSettingsNotFound
		}
		return payroll.PayrollSettings{}, fmt.Errorf("failed to get payroll settings: %w", err)
	}

	return s, nil
}

func (r *payrollRepository) UpsertSettings(ctx context.Context, settings payroll.PayrollSettings) (payroll.PayrollSettings, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO payroll_settings (
			company_id, late_deduction_enabled, late_deduction_per_minute,
			overtime_enabled, overtime_pay_per_minute,
			early_leave_deduction_enabled, early_leave_deduction_per_minute
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (company_id) DO UPDATE SET
			late_deduction_enabled = EXCLUDED.late_deduction_enabled,
			late_deduction_per_minute = EXCLUDED.late_deduction_per_minute,
			overtime_enabled = EXCLUDED.overtime_enabled,
			overtime_pay_per_minute = EXCLUDED.overtime_pay_per_minute,
			early_leave_deduction_enabled = EXCLUDED.early_leave_deduction_enabled,
			early_leave_deduction_per_minute = EXCLUDED.early_leave_deduction_per_minute,
			updated_at = NOW()
		RETURNING id, company_id, late_deduction_enabled, late_deduction_per_minute,
			overtime_enabled, overtime_pay_per_minute,
			early_leave_deduction_enabled, early_leave_deduction_per_minute,
			created_at, updated_at
	`

	var s payroll.PayrollSettings
	err := q.QueryRow(ctx, query,
		settings.CompanyID, settings.LateDeductionEnabled, settings.LateDeductionPerMinute,
		settings.OvertimeEnabled, settings.OvertimePayPerMinute,
		settings.EarlyLeaveDeductionEnabled, settings.EarlyLeaveDeductionPerMinute,
	).Scan(
		&s.ID, &s.CompanyID, &s.LateDeductionEnabled, &s.LateDeductionPerMinute,
		&s.OvertimeEnabled, &s.OvertimePayPerMinute,
		&s.EarlyLeaveDeductionEnabled, &s.EarlyLeaveDeductionPerMinute,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return payroll.PayrollSettings{}, fmt.Errorf("failed to upsert payroll settings: %w", err)
	}

	return s, nil
}

// ========== COMPONENTS ==========

func (r *payrollRepository) CreateComponent(ctx context.Context, component payroll.PayrollComponent) (payroll.PayrollComponent, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO payroll_components (company_id, name, type, description, is_taxable, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, company_id, name, type, description, is_taxable, is_active, created_at, updated_at
	`

	var c payroll.PayrollComponent
	err := q.QueryRow(ctx, query,
		component.CompanyID, component.Name, component.Type, component.Description, component.IsTaxable, component.IsActive,
	).Scan(
		&c.ID, &c.CompanyID, &c.Name, &c.Type, &c.Description, &c.IsTaxable, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uk_payroll_component_name") {
			return payroll.PayrollComponent{}, payroll.ErrPayrollComponentNameExists
		}
		return payroll.PayrollComponent{}, fmt.Errorf("failed to create payroll component: %w", err)
	}

	return c, nil
}

func (r *payrollRepository) GetComponentByID(ctx context.Context, id string, companyID string) (payroll.PayrollComponent, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, name, type, description, is_taxable, is_active, created_at, updated_at
		FROM payroll_components
		WHERE id = $1 AND company_id = $2
	`

	var c payroll.PayrollComponent
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&c.ID, &c.CompanyID, &c.Name, &c.Type, &c.Description, &c.IsTaxable, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.PayrollComponent{}, payroll.ErrPayrollComponentNotFound
		}
		return payroll.PayrollComponent{}, fmt.Errorf("failed to get payroll component: %w", err)
	}

	return c, nil
}

func (r *payrollRepository) GetComponentsByCompanyID(ctx context.Context, companyID string, activeOnly bool) ([]payroll.PayrollComponent, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, name, type, description, is_taxable, is_active, created_at, updated_at
		FROM payroll_components
		WHERE company_id = $1
	`
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY type, name"

	rows, err := q.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list payroll components: %w", err)
	}
	defer rows.Close()

	var components []payroll.PayrollComponent
	for rows.Next() {
		var c payroll.PayrollComponent
		if err := rows.Scan(
			&c.ID, &c.CompanyID, &c.Name, &c.Type, &c.Description, &c.IsTaxable, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan payroll component: %w", err)
		}
		components = append(components, c)
	}

	return components, nil
}

func (r *payrollRepository) UpdateComponent(ctx context.Context, companyID string, req payroll.UpdatePayrollComponentRequest) error {
	q := GetQuerier(ctx, r.db)

	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{req.ID, companyID}
	argIdx := 3

	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.IsTaxable != nil {
		setParts = append(setParts, fmt.Sprintf("is_taxable = $%d", argIdx))
		args = append(args, *req.IsTaxable)
		argIdx++
	}
	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	query := fmt.Sprintf(`
		UPDATE payroll_components
		SET %s
		WHERE id = $1 AND company_id = $2
		RETURNING id
	`, strings.Join(setParts, ", "))

	var updatedID string
	err := q.QueryRow(ctx, query, args...).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.ErrPayrollComponentNotFound
		}
		if strings.Contains(err.Error(), "uk_payroll_component_name") {
			return payroll.ErrPayrollComponentNameExists
		}
		return fmt.Errorf("failed to update payroll component: %w", err)
	}

	return nil
}

func (r *payrollRepository) DeleteComponent(ctx context.Context, id string, companyID string) error {
	q := GetQuerier(ctx, r.db)

	query := `DELETE FROM payroll_components WHERE id = $1 AND company_id = $2 RETURNING id`

	var deletedID string
	err := q.QueryRow(ctx, query, id, companyID).Scan(&deletedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.ErrPayrollComponentNotFound
		}
		return fmt.Errorf("failed to delete payroll component: %w", err)
	}

	return nil
}

// ========== EMPLOYEE COMPONENTS ==========

func (r *payrollRepository) AssignComponentToEmployee(ctx context.Context, assignment payroll.EmployeePayrollComponent, companyID string) (payroll.EmployeePayrollComponent, error) {
	q := GetQuerier(ctx, r.db)

	// Verify employee belongs to company
	var empExists bool
	err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL)`, assignment.EmployeeID, companyID).Scan(&empExists)
	if err != nil || !empExists {
		return payroll.EmployeePayrollComponent{}, payroll.ErrEmployeeNotFound
	}

	// Verify component belongs to company
	var compExists bool
	err = q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM payroll_components WHERE id = $1 AND company_id = $2)`, assignment.PayrollComponentID, companyID).Scan(&compExists)
	if err != nil || !compExists {
		return payroll.EmployeePayrollComponent{}, payroll.ErrPayrollComponentNotFound
	}

	query := `
		INSERT INTO employee_payroll_components (employee_id, payroll_component_id, amount, effective_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, employee_id, payroll_component_id, amount, effective_date, end_date, created_at, updated_at
	`

	var a payroll.EmployeePayrollComponent
	err = q.QueryRow(ctx, query,
		assignment.EmployeeID, assignment.PayrollComponentID, assignment.Amount, assignment.EffectiveDate, assignment.EndDate,
	).Scan(
		&a.ID, &a.EmployeeID, &a.PayrollComponentID, &a.Amount, &a.EffectiveDate, &a.EndDate, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return payroll.EmployeePayrollComponent{}, fmt.Errorf("failed to assign component to employee: %w", err)
	}

	return a, nil
}

func (r *payrollRepository) GetEmployeeComponents(ctx context.Context, employeeID string, companyID string, activeOnly bool) ([]payroll.EmployeePayrollComponent, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT epc.id, epc.employee_id, epc.payroll_component_id, epc.amount, 
			   epc.effective_date, epc.end_date, epc.created_at, epc.updated_at,
			   pc.name as component_name, pc.type as component_type
		FROM employee_payroll_components epc
		JOIN payroll_components pc ON epc.payroll_component_id = pc.id
		JOIN employees e ON epc.employee_id = e.id
		WHERE epc.employee_id = $1 AND e.company_id = $2
	`
	if activeOnly {
		query += ` AND epc.effective_date <= CURRENT_DATE AND (epc.end_date IS NULL OR epc.end_date >= CURRENT_DATE)`
	}
	query += " ORDER BY pc.type, pc.name"

	rows, err := q.Query(ctx, query, employeeID, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee components: %w", err)
	}
	defer rows.Close()

	var assignments []payroll.EmployeePayrollComponent
	for rows.Next() {
		var a payroll.EmployeePayrollComponent
		if err := rows.Scan(
			&a.ID, &a.EmployeeID, &a.PayrollComponentID, &a.Amount,
			&a.EffectiveDate, &a.EndDate, &a.CreatedAt, &a.UpdatedAt,
			&a.ComponentName, &a.ComponentType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan employee component: %w", err)
		}
		assignments = append(assignments, a)
	}

	return assignments, nil
}

func (r *payrollRepository) GetEmployeeComponentByID(ctx context.Context, id string, companyID string) (payroll.EmployeePayrollComponent, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT epc.id, epc.employee_id, epc.payroll_component_id, epc.amount, 
			   epc.effective_date, epc.end_date, epc.created_at, epc.updated_at,
			   pc.name as component_name, pc.type as component_type
		FROM employee_payroll_components epc
		JOIN payroll_components pc ON epc.payroll_component_id = pc.id
		JOIN employees e ON epc.employee_id = e.id
		WHERE epc.id = $1 AND e.company_id = $2
	`

	var a payroll.EmployeePayrollComponent
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&a.ID, &a.EmployeeID, &a.PayrollComponentID, &a.Amount,
		&a.EffectiveDate, &a.EndDate, &a.CreatedAt, &a.UpdatedAt,
		&a.ComponentName, &a.ComponentType,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.EmployeePayrollComponent{}, payroll.ErrEmployeeComponentNotFound
		}
		return payroll.EmployeePayrollComponent{}, fmt.Errorf("failed to get employee component: %w", err)
	}

	return a, nil
}

func (r *payrollRepository) UpdateEmployeeComponent(ctx context.Context, companyID string, req payroll.UpdateEmployeeComponentRequest) error {
	q := GetQuerier(ctx, r.db)

	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{req.ID, companyID}
	argIdx := 3

	if req.Amount != nil {
		setParts = append(setParts, fmt.Sprintf("amount = $%d", argIdx))
		args = append(args, *req.Amount)
		argIdx++
	}
	if req.EffectiveDate != nil {
		effectiveDate, err := time.Parse("2006-01-02", *req.EffectiveDate)
		if err == nil {
			setParts = append(setParts, fmt.Sprintf("effective_date = $%d", argIdx))
			args = append(args, effectiveDate)
			argIdx++
		}
	}
	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err == nil {
			setParts = append(setParts, fmt.Sprintf("end_date = $%d", argIdx))
			args = append(args, endDate)
			argIdx++
		}
	}

	query := fmt.Sprintf(`
		UPDATE employee_payroll_components epc
		SET %s
		FROM employees e
		WHERE epc.id = $1 AND epc.employee_id = e.id AND e.company_id = $2
		RETURNING epc.id
	`, strings.Join(setParts, ", "))

	var updatedID string
	err := q.QueryRow(ctx, query, args...).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.ErrEmployeeComponentNotFound
		}
		return fmt.Errorf("failed to update employee component: %w", err)
	}

	return nil
}

func (r *payrollRepository) RemoveEmployeeComponent(ctx context.Context, id string, companyID string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		DELETE FROM employee_payroll_components epc
		USING employees e
		WHERE epc.id = $1 AND epc.employee_id = e.id AND e.company_id = $2
		RETURNING epc.id
	`

	var deletedID string
	err := q.QueryRow(ctx, query, id, companyID).Scan(&deletedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.ErrEmployeeComponentNotFound
		}
		return fmt.Errorf("failed to remove employee component: %w", err)
	}

	return nil
}

// ========== PAYROLL RECORDS ==========

func (r *payrollRepository) CreatePayrollRecord(ctx context.Context, record payroll.PayrollRecord) (payroll.PayrollRecord, error) {
	q := GetQuerier(ctx, r.db)

	allowancesJSON, _ := json.Marshal(record.AllowancesDetail)
	deductionsJSON, _ := json.Marshal(record.DeductionsDetail)

	query := `
		INSERT INTO payroll_records (
			employee_id, company_id, period_month, period_year, base_salary,
			total_allowances, total_deductions, allowances_detail, deductions_detail,
			total_work_days, total_late_minutes, late_deduction_amount,
			total_early_leave_minutes, early_leave_deduction_amount,
			total_overtime_minutes, overtime_amount, gross_salary, net_salary, status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING id, employee_id, company_id, period_month, period_year, base_salary,
			total_allowances, total_deductions, allowances_detail, deductions_detail,
			total_work_days, total_late_minutes, late_deduction_amount,
			total_early_leave_minutes, early_leave_deduction_amount,
			total_overtime_minutes, overtime_amount, gross_salary, net_salary,
			status, paid_at, paid_by, notes, created_at, updated_at
	`

	var rec payroll.PayrollRecord
	var allowancesBytes, deductionsBytes []byte
	err := q.QueryRow(ctx, query,
		record.EmployeeID, record.CompanyID, record.PeriodMonth, record.PeriodYear, record.BaseSalary,
		record.TotalAllowances, record.TotalDeductions, allowancesJSON, deductionsJSON,
		record.TotalWorkDays, record.TotalLateMinutes, record.LateDeductionAmount,
		record.TotalEarlyLeaveMinutes, record.EarlyLeaveDeductionAmount,
		record.TotalOvertimeMinutes, record.OvertimeAmount, record.GrossSalary, record.NetSalary, record.Status, record.Notes,
	).Scan(
		&rec.ID, &rec.EmployeeID, &rec.CompanyID, &rec.PeriodMonth, &rec.PeriodYear, &rec.BaseSalary,
		&rec.TotalAllowances, &rec.TotalDeductions, &allowancesBytes, &deductionsBytes,
		&rec.TotalWorkDays, &rec.TotalLateMinutes, &rec.LateDeductionAmount,
		&rec.TotalEarlyLeaveMinutes, &rec.EarlyLeaveDeductionAmount,
		&rec.TotalOvertimeMinutes, &rec.OvertimeAmount, &rec.GrossSalary, &rec.NetSalary,
		&rec.Status, &rec.PaidAt, &rec.PaidBy, &rec.Notes, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uk_employee_period") {
			return payroll.PayrollRecord{}, payroll.ErrPayrollRecordAlreadyExists
		}
		return payroll.PayrollRecord{}, fmt.Errorf("failed to create payroll record: %w", err)
	}

	_ = json.Unmarshal(allowancesBytes, &rec.AllowancesDetail)
	_ = json.Unmarshal(deductionsBytes, &rec.DeductionsDetail)

	return rec, nil
}

func (r *payrollRepository) GetPayrollRecordByID(ctx context.Context, id string, companyID string) (payroll.PayrollRecord, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT pr.id, pr.employee_id, pr.company_id, pr.period_month, pr.period_year, pr.base_salary,
			   pr.total_allowances, pr.total_deductions, pr.allowances_detail, pr.deductions_detail,
			   pr.total_work_days, pr.total_late_minutes, pr.late_deduction_amount,
			   pr.total_early_leave_minutes, pr.early_leave_deduction_amount,
			   pr.total_overtime_minutes, pr.overtime_amount, pr.gross_salary, pr.net_salary,
			   pr.status, pr.paid_at, pr.paid_by, pr.notes, pr.created_at, pr.updated_at,
			   e.full_name as employee_name, e.employee_code, p.name as position_name, b.name as branch_name
		FROM payroll_records pr
		JOIN employees e ON pr.employee_id = e.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE pr.id = $1 AND pr.company_id = $2
	`

	var rec payroll.PayrollRecord
	var allowancesBytes, deductionsBytes []byte
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&rec.ID, &rec.EmployeeID, &rec.CompanyID, &rec.PeriodMonth, &rec.PeriodYear, &rec.BaseSalary,
		&rec.TotalAllowances, &rec.TotalDeductions, &allowancesBytes, &deductionsBytes,
		&rec.TotalWorkDays, &rec.TotalLateMinutes, &rec.LateDeductionAmount,
		&rec.TotalEarlyLeaveMinutes, &rec.EarlyLeaveDeductionAmount,
		&rec.TotalOvertimeMinutes, &rec.OvertimeAmount, &rec.GrossSalary, &rec.NetSalary,
		&rec.Status, &rec.PaidAt, &rec.PaidBy, &rec.Notes, &rec.CreatedAt, &rec.UpdatedAt,
		&rec.EmployeeName, &rec.EmployeeCode, &rec.PositionName, &rec.BranchName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.PayrollRecord{}, payroll.ErrPayrollRecordNotFound
		}
		return payroll.PayrollRecord{}, fmt.Errorf("failed to get payroll record: %w", err)
	}

	_ = json.Unmarshal(allowancesBytes, &rec.AllowancesDetail)
	_ = json.Unmarshal(deductionsBytes, &rec.DeductionsDetail)

	return rec, nil
}

func (r *payrollRepository) GetPayrollRecordByEmployeePeriod(ctx context.Context, employeeID string, month, year int, companyID string) (payroll.PayrollRecord, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT pr.id, pr.employee_id, pr.company_id, pr.period_month, pr.period_year, pr.base_salary,
			   pr.total_allowances, pr.total_deductions, pr.allowances_detail, pr.deductions_detail,
			   pr.total_work_days, pr.total_late_minutes, pr.late_deduction_amount,
			   pr.total_early_leave_minutes, pr.early_leave_deduction_amount,
			   pr.total_overtime_minutes, pr.overtime_amount, pr.gross_salary, pr.net_salary,
			   pr.status, pr.paid_at, pr.paid_by, pr.notes, pr.created_at, pr.updated_at
		FROM payroll_records pr
		JOIN employees e ON pr.employee_id = e.id
		WHERE pr.employee_id = $1 AND pr.period_month = $2 AND pr.period_year = $3 AND e.company_id = $4
	`

	var rec payroll.PayrollRecord
	var allowancesBytes, deductionsBytes []byte
	err := q.QueryRow(ctx, query, employeeID, month, year, companyID).Scan(
		&rec.ID, &rec.EmployeeID, &rec.CompanyID, &rec.PeriodMonth, &rec.PeriodYear, &rec.BaseSalary,
		&rec.TotalAllowances, &rec.TotalDeductions, &allowancesBytes, &deductionsBytes,
		&rec.TotalWorkDays, &rec.TotalLateMinutes, &rec.LateDeductionAmount,
		&rec.TotalEarlyLeaveMinutes, &rec.EarlyLeaveDeductionAmount,
		&rec.TotalOvertimeMinutes, &rec.OvertimeAmount, &rec.GrossSalary, &rec.NetSalary,
		&rec.Status, &rec.PaidAt, &rec.PaidBy, &rec.Notes, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.PayrollRecord{}, payroll.ErrPayrollRecordNotFound
		}
		return payroll.PayrollRecord{}, fmt.Errorf("failed to get payroll record: %w", err)
	}

	_ = json.Unmarshal(allowancesBytes, &rec.AllowancesDetail)
	_ = json.Unmarshal(deductionsBytes, &rec.DeductionsDetail)

	return rec, nil
}

func (r *payrollRepository) ListPayrollRecords(ctx context.Context, companyID string, filter payroll.PayrollFilter) ([]payroll.PayrollRecord, int64, error) {
	q := GetQuerier(ctx, r.db)

	baseQuery := `
		FROM payroll_records pr
		JOIN employees e ON pr.employee_id = e.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE pr.company_id = $1
	`
	args := []interface{}{companyID}
	argIdx := 2

	if filter.PeriodMonth != nil {
		baseQuery += fmt.Sprintf(" AND pr.period_month = $%d", argIdx)
		args = append(args, *filter.PeriodMonth)
		argIdx++
	}
	if filter.PeriodYear != nil {
		baseQuery += fmt.Sprintf(" AND pr.period_year = $%d", argIdx)
		args = append(args, *filter.PeriodYear)
		argIdx++
	}
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND pr.status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.EmployeeID != nil {
		baseQuery += fmt.Sprintf(" AND pr.employee_id = $%d", argIdx)
		args = append(args, *filter.EmployeeID)
		argIdx++
	}

	// Count query
	var totalCount int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count payroll records: %w", err)
	}

	// Sort
	sortColumn := "pr.created_at"
	if filter.SortBy != "" {
		allowedColumns := map[string]string{
			"created_at":    "pr.created_at",
			"period":        "pr.period_year DESC, pr.period_month",
			"employee_name": "e.full_name",
			"net_salary":    "pr.net_salary",
		}
		if col, ok := allowedColumns[filter.SortBy]; ok {
			sortColumn = col
		}
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Pagination
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	selectQuery := fmt.Sprintf(`
		SELECT pr.id, pr.employee_id, pr.company_id, pr.period_month, pr.period_year, pr.base_salary,
			   pr.total_allowances, pr.total_deductions, pr.allowances_detail, pr.deductions_detail,
			   pr.total_work_days, pr.total_late_minutes, pr.late_deduction_amount,
			   pr.total_early_leave_minutes, pr.early_leave_deduction_amount,
			   pr.total_overtime_minutes, pr.overtime_amount, pr.gross_salary, pr.net_salary,
			   pr.status, pr.paid_at, pr.paid_by, pr.notes, pr.created_at, pr.updated_at,
			   e.full_name as employee_name, e.employee_code, p.name as position_name, b.name as branch_name
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, baseQuery, sortColumn, sortOrder, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	rows, err := q.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payroll records: %w", err)
	}
	defer rows.Close()

	var records []payroll.PayrollRecord
	for rows.Next() {
		var rec payroll.PayrollRecord
		var allowancesBytes, deductionsBytes []byte
		if err := rows.Scan(
			&rec.ID, &rec.EmployeeID, &rec.CompanyID, &rec.PeriodMonth, &rec.PeriodYear, &rec.BaseSalary,
			&rec.TotalAllowances, &rec.TotalDeductions, &allowancesBytes, &deductionsBytes,
			&rec.TotalWorkDays, &rec.TotalLateMinutes, &rec.LateDeductionAmount,
			&rec.TotalEarlyLeaveMinutes, &rec.EarlyLeaveDeductionAmount,
			&rec.TotalOvertimeMinutes, &rec.OvertimeAmount, &rec.GrossSalary, &rec.NetSalary,
			&rec.Status, &rec.PaidAt, &rec.PaidBy, &rec.Notes, &rec.CreatedAt, &rec.UpdatedAt,
			&rec.EmployeeName, &rec.EmployeeCode, &rec.PositionName, &rec.BranchName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan payroll record: %w", err)
		}
		_ = json.Unmarshal(allowancesBytes, &rec.AllowancesDetail)
		_ = json.Unmarshal(deductionsBytes, &rec.DeductionsDetail)
		records = append(records, rec)
	}

	return records, totalCount, nil
}

func (r *payrollRepository) UpdatePayrollRecord(ctx context.Context, companyID string, req payroll.UpdatePayrollRecordRequest) error {
	q := GetQuerier(ctx, r.db)

	// Check if record is already paid
	var status string
	err := q.QueryRow(ctx, `SELECT status FROM payroll_records WHERE id = $1 AND company_id = $2`, req.ID, companyID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.ErrPayrollRecordNotFound
		}
		return fmt.Errorf("failed to check payroll record status: %w", err)
	}
	if status == string(payroll.PayrollStatusPaid) {
		return payroll.ErrPayrollRecordAlreadyPaid
	}

	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{req.ID, companyID}
	argIdx := 3

	if req.BaseSalary != nil {
		setParts = append(setParts, fmt.Sprintf("base_salary = $%d", argIdx))
		args = append(args, *req.BaseSalary)
		argIdx++
	}
	if req.TotalAllowances != nil {
		setParts = append(setParts, fmt.Sprintf("total_allowances = $%d", argIdx))
		args = append(args, *req.TotalAllowances)
		argIdx++
	}
	if req.TotalDeductions != nil {
		setParts = append(setParts, fmt.Sprintf("total_deductions = $%d", argIdx))
		args = append(args, *req.TotalDeductions)
		argIdx++
	}
	if req.LateDeductionAmount != nil {
		setParts = append(setParts, fmt.Sprintf("late_deduction_amount = $%d", argIdx))
		args = append(args, *req.LateDeductionAmount)
		argIdx++
	}
	if req.OvertimeAmount != nil {
		setParts = append(setParts, fmt.Sprintf("overtime_amount = $%d", argIdx))
		args = append(args, *req.OvertimeAmount)
		argIdx++
	}
	if req.Notes != nil {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *req.Notes)
		argIdx++
	}

	// Recalculate gross and net salary
	setParts = append(setParts, `
		gross_salary = COALESCE(base_salary, 0) + COALESCE(total_allowances, 0) + COALESCE(overtime_amount, 0),
		net_salary = COALESCE(base_salary, 0) + COALESCE(total_allowances, 0) + COALESCE(overtime_amount, 0) 
			- COALESCE(total_deductions, 0) - COALESCE(late_deduction_amount, 0) - COALESCE(early_leave_deduction_amount, 0)
	`)

	query := fmt.Sprintf(`
		UPDATE payroll_records
		SET %s
		WHERE id = $1 AND company_id = $2
		RETURNING id
	`, strings.Join(setParts, ", "))

	var updatedID string
	err = q.QueryRow(ctx, query, args...).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.ErrPayrollRecordNotFound
		}
		return fmt.Errorf("failed to update payroll record: %w", err)
	}

	return nil
}

func (r *payrollRepository) FinalizePayrollRecords(ctx context.Context, ids []string, paidBy string, companyID string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE payroll_records
		SET status = 'paid', paid_at = NOW(), paid_by = $1, updated_at = NOW()
		WHERE id = ANY($2) AND company_id = $3 AND status = 'draft'
	`

	_, err := q.Exec(ctx, query, paidBy, ids, companyID)
	if err != nil {
		return fmt.Errorf("failed to finalize payroll records: %w", err)
	}

	return nil
}

func (r *payrollRepository) DeletePayrollRecord(ctx context.Context, id string, companyID string) error {
	q := GetQuerier(ctx, r.db)

	// Check if record is already paid
	var status string
	err := q.QueryRow(ctx, `SELECT status FROM payroll_records WHERE id = $1 AND company_id = $2`, id, companyID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.ErrPayrollRecordNotFound
		}
		return fmt.Errorf("failed to check payroll record status: %w", err)
	}
	if status == string(payroll.PayrollStatusPaid) {
		return payroll.ErrCannotDeletePaidRecord
	}

	query := `DELETE FROM payroll_records WHERE id = $1 AND company_id = $2 RETURNING id`

	var deletedID string
	err = q.QueryRow(ctx, query, id, companyID).Scan(&deletedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return payroll.ErrPayrollRecordNotFound
		}
		return fmt.Errorf("failed to delete payroll record: %w", err)
	}

	return nil
}

// ========== AGGREGATIONS ==========

func (r *payrollRepository) GetAttendanceSummary(ctx context.Context, companyID string, month, year int, employeeIDs []string) ([]payroll.AttendanceSummary, error) {
	q := GetQuerier(ctx, r.db)

	// Note: Status can be 'on_time', 'late', or leave type names (dynamic).
	// Exclude: 'rejected', 'approved', 'absent', 'waiting_approval'
	query := `
		SELECT 
			employee_id,
			COUNT(*) as total_work_days,
			COALESCE(SUM(late_minutes), 0) as total_late_minutes,
			COALESCE(SUM(early_leave_minutes), 0) as total_early_leave_minutes,
			COALESCE(SUM(overtime_minutes), 0) as total_overtime_minutes
		FROM attendances
		WHERE company_id = $1 
			AND EXTRACT(MONTH FROM date) = $2
			AND EXTRACT(YEAR FROM date) = $3
			AND status NOT IN ('rejected', 'approved', 'absent', 'waiting_approval')
	`

	args := []interface{}{companyID, month, year}

	if len(employeeIDs) > 0 {
		query += ` AND employee_id = ANY($4)`
		args = append(args, employeeIDs)
	}

	query += ` GROUP BY employee_id`

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance summary: %w", err)
	}
	defer rows.Close()

	var summaries []payroll.AttendanceSummary
	for rows.Next() {
		var s payroll.AttendanceSummary
		if err := rows.Scan(
			&s.EmployeeID, &s.TotalWorkDays, &s.TotalLateMinutes,
			&s.TotalEarlyLeaveMinutes, &s.TotalOvertimeMinutes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attendance summary: %w", err)
		}
		summaries = append(summaries, s)
	}

	return summaries, nil
}

func (r *payrollRepository) GetPayrollSummary(ctx context.Context, companyID string, month, year int) (payroll.PayrollSummaryResponse, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			COUNT(*) as total_employees,
			COALESCE(SUM(base_salary), 0) as total_base_salary,
			COALESCE(SUM(total_allowances), 0) as total_allowances,
			COALESCE(SUM(total_deductions), 0) as total_deductions,
			COALESCE(SUM(late_deduction_amount), 0) as total_late_deduction,
			COALESCE(SUM(overtime_amount), 0) as total_overtime,
			COALESCE(SUM(gross_salary), 0) as total_gross_salary,
			COALESCE(SUM(net_salary), 0) as total_net_salary,
			COUNT(*) FILTER (WHERE status = 'draft') as draft_count,
			COUNT(*) FILTER (WHERE status = 'paid') as paid_count
		FROM payroll_records
		WHERE company_id = $1 AND period_month = $2 AND period_year = $3
	`

	var summary payroll.PayrollSummaryResponse
	err := q.QueryRow(ctx, query, companyID, month, year).Scan(
		&summary.TotalEmployees, &summary.TotalBaseSalary, &summary.TotalAllowances,
		&summary.TotalDeductions, &summary.TotalLateDeduction, &summary.TotalOvertime,
		&summary.TotalGrossSalary, &summary.TotalNetSalary, &summary.DraftCount, &summary.PaidCount,
	)
	if err != nil {
		return payroll.PayrollSummaryResponse{}, fmt.Errorf("failed to get payroll summary: %w", err)
	}

	summary.PeriodMonth = month
	summary.PeriodYear = year

	return summary, nil
}

// Helper to convert decimal map to string for JSON storage (for future use)
func decimalMapToStringMap(m map[string]decimal.Decimal) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v.String()
	}
	return result
}
