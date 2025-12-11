package dashboard

import (
	"context"
	"time"
)

// EmployeeSummaryStats combines all employee summary counts in single query
type EmployeeSummaryStats struct {
	Total    int64
	New      int64 // hired within 30 days
	Active   int64
	Resigned int64
}

// EmployeeMonthlyStats combines new/active/resign counts for a month
type EmployeeMonthlyStats struct {
	New    int64
	Active int64
	Resign int64
}

// EmployeeTypeStats combines all employment type counts
type EmployeeTypeStats struct {
	Permanent  int64
	Probation  int64
	Contract   int64
	Internship int64
	Freelance  int64
}

// AttendanceStats combines on_time/late/absent counts
type AttendanceStats struct {
	OnTime int64
	Late   int64
	Absent int64
}

// MonthlyAttendanceData combines attendance counts with latest records
type MonthlyAttendanceData struct {
	OnTime  int64
	Late    int64
	Absent  int64
	Records []AttendanceRecordItem
}

// DashboardRepository defines the interface for dashboard data access
type DashboardRepository interface {
	// GetEmployeeSummary returns total, new (30 days), active, resigned counts in single query
	GetEmployeeSummary(ctx context.Context, companyID string, since time.Time) (*EmployeeSummaryStats, error)

	// GetEmployeeMonthlyStats returns new/active/resign for a specific month in single query
	GetEmployeeMonthlyStats(ctx context.Context, companyID string, year, month int) (*EmployeeMonthlyStats, error)

	// GetEmployeeTypeStats returns employment type distribution in single query
	GetEmployeeTypeStats(ctx context.Context, companyID string, year, month int) (*EmployeeTypeStats, error)

	// GetAttendanceStatsByDay returns on_time/late/absent for a day
	GetAttendanceStatsByDay(ctx context.Context, companyID string, date time.Time) (*AttendanceStats, error)

	// GetMonthlyAttendanceWithRecords returns monthly stats + latest records in single query (using subquery)
	GetMonthlyAttendanceWithRecords(ctx context.Context, companyID string, year, month int, limit int) (*MonthlyAttendanceData, error)
}
