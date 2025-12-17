package report

import (
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

// ========================================
// MONTHLY ATTENDANCE REPORT
// ========================================

type MonthlyAttendanceReportRequest struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

func (r *MonthlyAttendanceReportRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.Month < 1 || r.Month > 12 {
		errs = append(errs, validator.ValidationError{
			Field:   "month",
			Message: "month must be between 1 and 12",
		})
	}

	currentYear := time.Now().Year()
	if r.Year < 2020 || r.Year > currentYear+1 {
		errs = append(errs, validator.ValidationError{
			Field:   "year",
			Message: fmt.Sprintf("year must be between 2020 and %d", currentYear+1),
		})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type MonthlyAttendanceReport struct {
	PeriodMonth int    `json:"period_month"`
	PeriodYear  int    `json:"period_year"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	GeneratedAt string `json:"generated_at"`

	Employees []MonthlyAttendanceEmployee `json:"employees"`
}

type MonthlyAttendanceEmployee struct {
	EmployeeID   string `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	EmployeeNIK  string `json:"employee_nik"`
	Position     string `json:"position"`

	Summary   AttendanceSummary    `json:"summary"`
	DailyLogs []AttendanceDailyLog `json:"daily_logs"`
}

type AttendanceSummary struct {
	TotalWorkDays    int     `json:"total_work_days"`
	TotalWorkHours   float64 `json:"total_work_hours"`
	TotalLateMinutes int     `json:"total_late_minutes"`
	TotalPresent     int     `json:"total_present"`
	TotalLeave       int     `json:"total_leave"`
	TotalLateDays    int     `json:"total_late_days"`
}

type AttendanceDailyLog struct {
	Date              string  `json:"date"`
	DayOfWeek         string  `json:"day_of_week"`
	ShiftName         string  `json:"shift_name"`
	ClockIn           *string `json:"clock_in"`
	ClockOut          *string `json:"clock_out"`
	Status            string  `json:"status"`
	LateMinutes       int     `json:"late_minutes"`
	EarlyLeaveMinutes int     `json:"early_leave_minutes"`
}

// ========================================
// EMPLOYEE PAYROLL SUMMARY
// ========================================

type PayrollSummaryReportRequest struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

func (r *PayrollSummaryReportRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.Month < 1 || r.Month > 12 {
		errs = append(errs, validator.ValidationError{
			Field:   "month",
			Message: "month must be between 1 and 12",
		})
	}

	currentYear := time.Now().Year()
	if r.Year < 2020 || r.Year > currentYear+1 {
		errs = append(errs, validator.ValidationError{
			Field:   "year",
			Message: fmt.Sprintf("year must be between 2020 and %d", currentYear+1),
		})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type PayrollSummaryReport struct {
	PeriodMonth      int     `json:"period_month"`
	PeriodYear       int     `json:"period_year"`
	GeneratedAt      string  `json:"generated_at"`
	TotalGrossPayout float64 `json:"total_gross_payout"`
	TotalNetPayout   float64 `json:"total_net_payout"`
	TotalEmployees   int     `json:"total_employees"`

	Rows []PayrollSummaryRow `json:"rows"`
}

type PayrollSummaryRow struct {
	EmployeeName string `json:"employee_name"`
	EmployeeCode string `json:"employee_code"`
	Position     string `json:"position"`

	// Earnings
	BaseSalary      float64 `json:"base_salary"`
	TotalAllowances float64 `json:"total_allowances"`
	OvertimeAmount  float64 `json:"overtime_amount"`
	GrossSalary     float64 `json:"gross_salary"`

	// Deductions
	LateDeduction       float64 `json:"late_deduction"`
	EarlyLeaveDeduction float64 `json:"early_leave_deduction"`
	OtherDeductions     float64 `json:"other_deductions"`

	// Final
	NetSalary float64 `json:"net_salary"`
}

// ========================================
// LEAVE BALANCE REPORT
// ========================================

type LeaveBalanceReportRequest struct {
	Year int `json:"year"`
}

func (r *LeaveBalanceReportRequest) Validate() error {
	var errs validator.ValidationErrors

	currentYear := time.Now().Year()
	if r.Year < 2020 || r.Year > currentYear+1 {
		errs = append(errs, validator.ValidationError{
			Field:   "year",
			Message: fmt.Sprintf("year must be between 2020 and %d", currentYear+1),
		})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type LeaveBalanceReport struct {
	GeneratedAt string `json:"generated_at"`
	Year        int    `json:"year"`

	Rows []LeaveBalanceRow `json:"rows"`
}

type LeaveBalanceRow struct {
	EmployeeID   string `json:"employee_id"`
	EmployeeCode string `json:"employee_code"`
	FullName     string `json:"full_name"`
	Position     string `json:"position"`
	JoinDate     string `json:"join_date"`

	Balances []LeaveTypeBalance `json:"balances"`
}

type LeaveTypeBalance struct {
	LeaveTypeName  string  `json:"leave_type_name"`
	LeaveTypeCode  string  `json:"leave_type_code"`
	TotalQuota     float64 `json:"total_quota"`
	UsedQuota      float64 `json:"used_quota"`
	PendingQuota   float64 `json:"pending_quota"`
	AvailableQuota float64 `json:"available_quota"`
}

// ========================================
// NEW HIRE REPORT
// ========================================

type NewHireReportRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (r *NewHireReportRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.StartDate == "" {
		errs = append(errs, validator.ValidationError{
			Field:   "start_date",
			Message: "start_date is required",
		})
	}

	if r.EndDate == "" {
		errs = append(errs, validator.ValidationError{
			Field:   "end_date",
			Message: "end_date is required",
		})
	}

	// Validate date format and order
	if r.StartDate != "" && r.EndDate != "" {
		startDate, err := time.Parse("2006-01-02", r.StartDate)
		if err != nil {
			errs = append(errs, validator.ValidationError{
				Field:   "start_date",
				Message: "start_date must be in YYYY-MM-DD format",
			})
		}

		endDate, err := time.Parse("2006-01-02", r.EndDate)
		if err != nil {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must be in YYYY-MM-DD format",
			})
		}

		if startDate.After(endDate) {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must be after start_date",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type NewHireReport struct {
	GeneratedAt string `json:"generated_at"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`

	Rows []NewHireRow `json:"rows"`
}

type NewHireRow struct {
	EmployeeCode   string `json:"employee_code"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	Position       string `json:"position"`
	Grade          string `json:"grade"`
	Branch         string `json:"branch"`
	HireDate       string `json:"hire_date"`
	EmploymentType string `json:"employment_type"`
	SystemStatus   string `json:"system_status"`
}
