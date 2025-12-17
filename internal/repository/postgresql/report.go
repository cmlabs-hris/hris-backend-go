package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/report"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type ReportRepository interface {
	GetMonthlyAttendanceReport(ctx context.Context, companyID string, month, year int) ([]report.MonthlyAttendanceEmployee, error)
	GetPayrollSummaryReport(ctx context.Context, companyID string, month, year int) ([]report.PayrollSummaryRow, error)
	GetLeaveBalanceReport(ctx context.Context, companyID string, year int) ([]report.LeaveBalanceRow, error)
	GetNewHireReport(ctx context.Context, companyID, startDate, endDate string) ([]report.NewHireRow, error)
}

type reportRepositoryImpl struct {
	db *database.DB
}

func NewReportRepository(db *database.DB) ReportRepository {
	return &reportRepositoryImpl{db: db}
}

// GetMonthlyAttendanceReport retrieves attendance data for all employees in a company for a specific month
func (r *reportRepositoryImpl) GetMonthlyAttendanceReport(ctx context.Context, companyID string, month, year int) ([]report.MonthlyAttendanceEmployee, error) {
	q := GetQuerier(ctx, r.db)

	// Calculate period start and end dates
	periodStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	periodEnd := periodStart.AddDate(0, 1, -1)

	// Complex SQL query with CTE for efficient data retrieval
	query := `
		WITH employee_attendance AS (
			SELECT 
				e.id as employee_id,
				e.full_name as employee_name,
				e.employee_code as employee_nik,
				p.name as position_name,
				a.date,
				a.clock_in,
				a.clock_out,
				a.work_hours_in_minutes,
				a.status,
				a.late_minutes,
				a.early_leave_minutes,
				a.leave_type_id,
				COALESCE(ws.name, 'Default') as shift_name
			FROM employees e
			JOIN positions p ON e.position_id = p.id
			LEFT JOIN attendances a ON e.id = a.employee_id 
				AND a.date >= $2 AND a.date <= $3
			LEFT JOIN work_schedule_times wst ON a.work_schedule_time_id = wst.id
			LEFT JOIN work_schedules ws ON wst.work_schedule_id = ws.id
			WHERE e.company_id = $1 
				AND e.deleted_at IS NULL
				AND e.employment_status = 'active'
		),
		employee_summary AS (
			SELECT 
				employee_id,
				employee_name,
				employee_nik,
				position_name,
				COUNT(CASE WHEN date IS NOT NULL AND status != 'absent' THEN 1 END) as total_work_days,
				COALESCE(SUM(work_hours_in_minutes), 0) / 60.0 as total_work_hours,
				COALESCE(SUM(late_minutes), 0) as total_late_minutes,
				COUNT(CASE WHEN status = 'present' OR status = 'approved' THEN 1 END) as total_present,
				COUNT(CASE WHEN leave_type_id IS NOT NULL THEN 1 END) as total_leave,
				COUNT(CASE WHEN late_minutes > 0 THEN 1 END) as total_late_days
			FROM employee_attendance
			GROUP BY employee_id, employee_name, employee_nik, position_name
		)
		SELECT 
			es.employee_id,
			es.employee_name,
			es.employee_nik,
			es.position_name,
			es.total_work_days,
			es.total_work_hours,
			es.total_late_minutes,
			es.total_present,
			es.total_leave,
			es.total_late_days
		FROM employee_summary es
		ORDER BY es.employee_name ASC
	`

	rows, err := q.Query(ctx, query, companyID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to query attendance summary: %w", err)
	}
	defer rows.Close()

	// Map to store employees
	employeeMap := make(map[string]*report.MonthlyAttendanceEmployee)
	var employeeOrder []string

	for rows.Next() {
		var emp report.MonthlyAttendanceEmployee
		var totalWorkHours float64

		err := rows.Scan(
			&emp.EmployeeID,
			&emp.EmployeeName,
			&emp.EmployeeNIK,
			&emp.Position,
			&emp.Summary.TotalWorkDays,
			&totalWorkHours,
			&emp.Summary.TotalLateMinutes,
			&emp.Summary.TotalPresent,
			&emp.Summary.TotalLeave,
			&emp.Summary.TotalLateDays,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance summary: %w", err)
		}

		emp.Summary.TotalWorkHours = totalWorkHours
		emp.DailyLogs = []report.AttendanceDailyLog{}
		employeeMap[emp.EmployeeID] = &emp
		employeeOrder = append(employeeOrder, emp.EmployeeID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Get daily logs for all employees
	dailyLogsQuery := `
		SELECT 
			e.id as employee_id,
			a.date,
			COALESCE(ws.name, 'Default') as shift_name,
			a.clock_in,
			a.clock_out,
			a.status,
			COALESCE(a.late_minutes, 0) as late_minutes,
			COALESCE(a.early_leave_minutes, 0) as early_leave_minutes
		FROM employees e
		LEFT JOIN attendances a ON e.id = a.employee_id 
			AND a.date >= $2 AND a.date <= $3
		LEFT JOIN work_schedule_times wst ON a.work_schedule_time_id = wst.id
		LEFT JOIN work_schedules ws ON wst.work_schedule_id = ws.id
		WHERE e.company_id = $1 
			AND e.deleted_at IS NULL
			AND e.employment_status = 'active'
			AND a.date IS NOT NULL
		ORDER BY e.id, a.date ASC
	`

	logRows, err := q.Query(ctx, dailyLogsQuery, companyID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily logs: %w", err)
	}
	defer logRows.Close()

	for logRows.Next() {
		var employeeID string
		var log report.AttendanceDailyLog
		var date time.Time
		var clockIn, clockOut *time.Time

		err := logRows.Scan(
			&employeeID,
			&date,
			&log.ShiftName,
			&clockIn,
			&clockOut,
			&log.Status,
			&log.LateMinutes,
			&log.EarlyLeaveMinutes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily log: %w", err)
		}

		log.Date = date.Format("2006-01-02")
		log.DayOfWeek = date.Weekday().String()

		if clockIn != nil {
			ci := clockIn.Format("15:04")
			log.ClockIn = &ci
		}
		if clockOut != nil {
			co := clockOut.Format("15:04")
			log.ClockOut = &co
		}

		if emp, ok := employeeMap[employeeID]; ok {
			emp.DailyLogs = append(emp.DailyLogs, log)
		}
	}

	if err := logRows.Err(); err != nil {
		return nil, fmt.Errorf("log rows error: %w", err)
	}

	// Build result in order
	result := make([]report.MonthlyAttendanceEmployee, 0, len(employeeOrder))
	for _, id := range employeeOrder {
		if emp, ok := employeeMap[id]; ok {
			result = append(result, *emp)
		}
	}

	return result, nil
}

// GetPayrollSummaryReport retrieves payroll data for all employees in a company for a specific month
func (r *reportRepositoryImpl) GetPayrollSummaryReport(ctx context.Context, companyID string, month, year int) ([]report.PayrollSummaryRow, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			e.full_name as employee_name,
			e.employee_code,
			p.name as position_name,
			COALESCE(pr.base_salary, 0) as base_salary,
			COALESCE(pr.total_allowances, 0) as total_allowances,
			COALESCE(pr.overtime_amount, 0) as overtime_amount,
			COALESCE(pr.base_salary, 0) + COALESCE(pr.total_allowances, 0) + COALESCE(pr.overtime_amount, 0) as gross_salary,
			COALESCE(pr.late_deduction_amount, 0) as late_deduction,
			COALESCE(pr.early_leave_deduction_amount, 0) as early_leave_deduction,
			COALESCE(pr.total_deductions, 0) - COALESCE(pr.late_deduction_amount, 0) - COALESCE(pr.early_leave_deduction_amount, 0) as other_deductions,
			COALESCE(pr.net_salary, 0) as net_salary
		FROM employees e
		JOIN positions p ON e.position_id = p.id
		LEFT JOIN payroll_records pr ON e.id = pr.employee_id 
			AND pr.period_month = $2 
			AND pr.period_year = $3
		WHERE e.company_id = $1 
			AND e.deleted_at IS NULL
			AND e.employment_status = 'active'
		ORDER BY e.full_name ASC
	`

	rows, err := q.Query(ctx, query, companyID, month, year)
	if err != nil {
		return nil, fmt.Errorf("failed to query payroll summary: %w", err)
	}
	defer rows.Close()

	var result []report.PayrollSummaryRow

	for rows.Next() {
		var row report.PayrollSummaryRow

		err := rows.Scan(
			&row.EmployeeName,
			&row.EmployeeCode,
			&row.Position,
			&row.BaseSalary,
			&row.TotalAllowances,
			&row.OvertimeAmount,
			&row.GrossSalary,
			&row.LateDeduction,
			&row.EarlyLeaveDeduction,
			&row.OtherDeductions,
			&row.NetSalary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payroll row: %w", err)
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil
}

// GetLeaveBalanceReport retrieves leave balance for all employees in a company for a specific year
func (r *reportRepositoryImpl) GetLeaveBalanceReport(ctx context.Context, companyID string, year int) ([]report.LeaveBalanceRow, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			e.id as employee_id,
			e.employee_code,
			e.full_name,
			p.name as position_name,
			e.hire_date,
			lt.name as leave_name,
			COALESCE(lt.code, '') as leave_code,
			COALESCE(lq.opening_balance, 0) + COALESCE(lq.earned_quota, 0) + COALESCE(lq.rollover_quota, 0) + COALESCE(lq.adjustment_quota, 0) as total_quota,
			COALESCE(lq.used_quota, 0) as used_quota,
			COALESCE(lq.pending_quota, 0) as pending_quota,
			COALESCE(lq.available_quota, 0) as available_quota
		FROM employees e
		JOIN positions p ON e.position_id = p.id
		LEFT JOIN leave_quotas lq ON e.id = lq.employee_id AND lq.year = $2
		LEFT JOIN leave_types lt ON lq.leave_type_id = lt.id
		WHERE 
			e.company_id = $1
			AND e.employment_status = 'active'
			AND e.deleted_at IS NULL
		ORDER BY e.full_name ASC, lt.name ASC
	`

	rows, err := q.Query(ctx, query, companyID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to query leave balance: %w", err)
	}
	defer rows.Close()

	// Map to aggregate leave balances per employee
	employeeMap := make(map[string]*report.LeaveBalanceRow)
	var employeeOrder []string

	for rows.Next() {
		var employeeID, employeeCode, fullName, positionName string
		var hireDate time.Time
		var leaveName, leaveCode *string
		var totalQuota, usedQuota, pendingQuota, availableQuota *float64

		err := rows.Scan(
			&employeeID,
			&employeeCode,
			&fullName,
			&positionName,
			&hireDate,
			&leaveName,
			&leaveCode,
			&totalQuota,
			&usedQuota,
			&pendingQuota,
			&availableQuota,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave balance: %w", err)
		}

		// Get or create employee entry
		emp, exists := employeeMap[employeeID]
		if !exists {
			emp = &report.LeaveBalanceRow{
				EmployeeID:   employeeID,
				EmployeeCode: employeeCode,
				FullName:     fullName,
				Position:     positionName,
				JoinDate:     hireDate.Format("2006-01-02"),
				Balances:     []report.LeaveTypeBalance{},
			}
			employeeMap[employeeID] = emp
			employeeOrder = append(employeeOrder, employeeID)
		}

		// Add leave balance if exists
		if leaveName != nil && *leaveName != "" {
			balance := report.LeaveTypeBalance{
				LeaveTypeName: *leaveName,
			}
			if leaveCode != nil {
				balance.LeaveTypeCode = *leaveCode
			}
			if totalQuota != nil {
				balance.TotalQuota = *totalQuota
			}
			if usedQuota != nil {
				balance.UsedQuota = *usedQuota
			}
			if pendingQuota != nil {
				balance.PendingQuota = *pendingQuota
			}
			if availableQuota != nil {
				balance.AvailableQuota = *availableQuota
			}
			emp.Balances = append(emp.Balances, balance)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Build result in order
	result := make([]report.LeaveBalanceRow, 0, len(employeeOrder))
	for _, id := range employeeOrder {
		if emp, ok := employeeMap[id]; ok {
			result = append(result, *emp)
		}
	}

	return result, nil
}

// GetNewHireReport retrieves new hires within a date range
func (r *reportRepositoryImpl) GetNewHireReport(ctx context.Context, companyID, startDate, endDate string) ([]report.NewHireRow, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			e.employee_code,
			e.full_name,
			COALESCE(u.email, inv.email, 'No Email') as email_address,
			p.name as position_name,
			g.name as grade_name,
			b.name as branch_name,
			e.hire_date,
			e.employment_type::text,
			CASE 
				WHEN u.id IS NOT NULL THEN 'Active User'
				WHEN inv.status = 'pending' THEN 'Invitation Sent'
				ELSE 'Not Invited'
			END as system_status
		FROM employees e
		JOIN positions p ON e.position_id = p.id
		JOIN grades g ON e.grade_id = g.id
		JOIN branches b ON e.branch_id = b.id
		LEFT JOIN users u ON e.user_id = u.id
		LEFT JOIN employee_invitations inv ON e.id = inv.employee_id AND inv.status = 'pending'
		WHERE 
			e.company_id = $1
			AND e.hire_date BETWEEN $2 AND $3
			AND e.deleted_at IS NULL
		ORDER BY e.hire_date DESC
	`

	rows, err := q.Query(ctx, query, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query new hires: %w", err)
	}
	defer rows.Close()

	var result []report.NewHireRow

	for rows.Next() {
		var row report.NewHireRow
		var hireDate time.Time

		err := rows.Scan(
			&row.EmployeeCode,
			&row.FullName,
			&row.Email,
			&row.Position,
			&row.Grade,
			&row.Branch,
			&hireDate,
			&row.EmploymentType,
			&row.SystemStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan new hire: %w", err)
		}

		row.HireDate = hireDate.Format("2006-01-02")
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil
}
