package employee_dashboard

import "context"

// EmployeeDashboardService defines the interface for employee dashboard operations
type EmployeeDashboardService interface {
	// GetDashboard returns combined employee dashboard data
	GetDashboard(ctx context.Context) (*EmployeeDashboardResponse, error)

	// GetWorkStats returns work stats for a date range (default: current month)
	// startDate, endDate format: "YYYY-MM-DD"
	GetWorkStats(ctx context.Context, startDate, endDate string) (*WorkStatsResponse, error)

	// GetAttendanceSummary returns attendance summary for a month
	// month format: "YYYY-MM"
	GetAttendanceSummary(ctx context.Context, month string) (*AttendanceSummaryResponse, error)

	// GetLeaveSummary returns leave quota summary for a year
	GetLeaveSummary(ctx context.Context, year string) (*LeaveSummaryResponse, error)

	// GetWorkHoursChart returns daily work hours for a specific week
	// week: 1, 2, 3, 4, etc (default: current week of month)
	GetWorkHoursChart(ctx context.Context, week string) (*WorkHoursChartResponse, error)
}
