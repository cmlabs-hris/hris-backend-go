package dashboard

import "context"

// DashboardService defines the interface for dashboard operations
type DashboardService interface {
	// GetDashboard returns combined dashboard data using goroutines
	GetDashboard(ctx context.Context) (*DashboardResponse, error)

	// GetEmployeeCurrentNumber returns employee counts (new/active/resign) for a month
	GetEmployeeCurrentNumber(ctx context.Context, month string) (*EmployeeCurrentNumberResponse, error)

	// GetEmployeeStatusStats returns employee distribution by employment type for a month
	GetEmployeeStatusStats(ctx context.Context, month string) (*EmployeeStatusStatsResponse, error)

	// GetMonthlyAttendance returns monthly attendance summary with latest 10 records
	GetMonthlyAttendance(ctx context.Context, month string) (*MonthlyAttendanceResponse, error)

	// GetDailyAttendanceStats returns attendance statistics for a specific day
	GetDailyAttendanceStats(ctx context.Context, date string) (*AttendanceStatsResponse, error)
}
