package employee_dashboard

import (
	"context"
	"time"
)

// EmployeeDashboardRepository defines the interface for employee dashboard data access
type EmployeeDashboardRepository interface {
	// GetWorkStats returns work hours and attendance counts for a date range
	GetWorkStats(ctx context.Context, employeeID string, startDate, endDate time.Time) (*WorkStatsData, error)

	// GetAttendanceSummary returns attendance distribution for a month
	GetAttendanceSummary(ctx context.Context, employeeID string, year, month int) (*AttendanceSummaryData, error)

	// GetLeaveSummary returns leave quota summary for a year
	GetLeaveSummary(ctx context.Context, employeeID string, year int) ([]LeaveQuotaData, error)

	// GetWorkHoursChart returns daily work hours for a specific week
	GetWorkHoursChart(ctx context.Context, employeeID string, year, month, week int) ([]DailyWorkHourData, error)
}

// WorkStatsData contains raw work stats from DB
type WorkStatsData struct {
	TotalWorkMinutes int64
	OnTimeCount      int64
	LateCount        int64
	AbsentCount      int64
}

// AttendanceSummaryData contains raw attendance summary from DB
type AttendanceSummaryData struct {
	OnTime         int64
	Late           int64
	Absent         int64
	LeaveCount     int64
	LeaveBreakdown []LeaveBreakdownData
}

// LeaveBreakdownData contains leave count by type
type LeaveBreakdownData struct {
	LeaveTypeID   string
	LeaveTypeName string
	Count         int64
}

// LeaveQuotaData contains leave quota info from DB
type LeaveQuotaData struct {
	LeaveTypeID    string
	LeaveTypeName  string
	TotalQuota     float64 // earned_quota + rollover_quota + adjustment_quota
	UsedQuota      float64
	AvailableQuota float64
}

// DailyWorkHourData contains work hours for a single day
type DailyWorkHourData struct {
	Date        time.Time
	WorkMinutes int64
}
