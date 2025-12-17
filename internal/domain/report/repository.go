package report

import "context"

// ReportRepository defines the interface for report data access
type ReportRepository interface {
	// Monthly Attendance Report
	GetMonthlyAttendanceReport(ctx context.Context, companyID string, month, year int) ([]MonthlyAttendanceEmployee, error)

	// Payroll Summary Report
	GetPayrollSummaryReport(ctx context.Context, companyID string, month, year int) ([]PayrollSummaryRow, error)

	// Leave Balance Report
	GetLeaveBalanceReport(ctx context.Context, companyID string, year int) ([]LeaveBalanceRow, error)

	// New Hire Report
	GetNewHireReport(ctx context.Context, companyID, startDate, endDate string) ([]NewHireRow, error)
}
