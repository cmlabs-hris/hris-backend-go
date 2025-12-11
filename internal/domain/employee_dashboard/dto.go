package employee_dashboard

// ========== COMBINED EMPLOYEE DASHBOARD ==========

// EmployeeDashboardResponse is the combined response for employee dashboard
type EmployeeDashboardResponse struct {
	WorkStats         WorkStatsResponse         `json:"work_stats"`
	AttendanceSummary AttendanceSummaryResponse `json:"attendance_summary"`
	LeaveSummary      LeaveSummaryResponse      `json:"leave_summary"`
	WorkHoursChart    WorkHoursChartResponse    `json:"work_hours_chart"`
}

// ========== WORK STATS (Top Cards) ==========

// WorkStatsResponse contains work hours and attendance counts for date range
type WorkStatsResponse struct {
	WorkHours   string `json:"work_hours"`    // Format: "120h 54m"
	WorkMinutes int64  `json:"work_minutes"`  // Total minutes for calculation
	OnTimeCount int64  `json:"on_time_count"` // Count of on_time attendance
	LateCount   int64  `json:"late_count"`    // Count of late attendance
	AbsentCount int64  `json:"absent_count"`  // Count of absent attendance
	StartDate   string `json:"start_date"`    // Filter start date
	EndDate     string `json:"end_date"`      // Filter end date
}

// ========== ATTENDANCE SUMMARY (Pie Chart) ==========

// AttendanceSummaryResponse represents attendance distribution for a month
type AttendanceSummaryResponse struct {
	TotalAttendance int64                `json:"total_attendance"`
	OnTime          int64                `json:"on_time"`
	Late            int64                `json:"late"`
	Absent          int64                `json:"absent"`
	LeaveCount      int64                `json:"leave_count"` // Total leave attendance
	OnTimePercent   float64              `json:"on_time_percent"`
	LatePercent     float64              `json:"late_percent"`
	AbsentPercent   float64              `json:"absent_percent"`
	LeavePercent    float64              `json:"leave_percent"`
	LeaveBreakdown  []LeaveBreakdownItem `json:"leave_breakdown"` // Breakdown by leave type
	Month           string               `json:"month"`           // Format: "YYYY-MM"
}

// LeaveBreakdownItem represents leave count by type
type LeaveBreakdownItem struct {
	LeaveTypeName string  `json:"leave_type_name"`
	Count         int64   `json:"count"`
	Percent       float64 `json:"percent"`
}

// ========== LEAVE SUMMARY ==========

// LeaveSummaryResponse represents leave quota summary for a year
type LeaveSummaryResponse struct {
	Year             int              `json:"year"`
	LeaveQuotaDetail []LeaveQuotaItem `json:"leave_quota_detail"`
}

// LeaveQuotaItem represents quota info for a leave type
type LeaveQuotaItem struct {
	LeaveTypeID   string  `json:"leave_type_id"`
	LeaveTypeName string  `json:"leave_type_name"`
	TotalQuota    float64 `json:"total_quota"` // Total earned quota
	Taken         float64 `json:"taken"`       // Used/approved leave days
	Remaining     float64 `json:"remaining"`   // Available quota
}

// ========== WORK HOURS CHART (Bar Chart) ==========

// WorkHoursChartResponse represents daily work hours for a week
type WorkHoursChartResponse struct {
	TotalWorkHours   string              `json:"total_work_hours"`   // Format: "120h 54m"
	TotalWorkMinutes int64               `json:"total_work_minutes"` // Total minutes
	WeekNumber       int                 `json:"week_number"`        // 1, 2, 3, 4, etc
	Year             int                 `json:"year"`
	Month            int                 `json:"month"`
	DailyWorkHours   []DailyWorkHourItem `json:"daily_work_hours"`
}

// DailyWorkHourItem represents work hours for a single day
type DailyWorkHourItem struct {
	Date        string `json:"date"`         // Format: "2006-01-02"
	DayName     string `json:"day_name"`     // "Monday", "Tuesday", etc
	WorkHours   string `json:"work_hours"`   // Format: "8h 30m"
	WorkMinutes int64  `json:"work_minutes"` // Total minutes
}
