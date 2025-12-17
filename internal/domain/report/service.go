package report

import "context"

// ReportService defines the interface for report generation
type ReportService interface {
	// Generate Monthly Attendance Report
	GenerateMonthlyAttendanceReport(ctx context.Context, req MonthlyAttendanceReportRequest) (MonthlyAttendanceReport, error)

	// Generate Payroll Summary Report
	GeneratePayrollSummaryReport(ctx context.Context, req PayrollSummaryReportRequest) (PayrollSummaryReport, error)

	// Generate Leave Balance Report
	GenerateLeaveBalanceReport(ctx context.Context, req LeaveBalanceReportRequest) (LeaveBalanceReport, error)

	// Generate New Hire Report
	GenerateNewHireReport(ctx context.Context, req NewHireReportRequest) (NewHireReport, error)
}
