package dashboard

// ========== COMBINED DASHBOARD ==========

// DashboardResponse is the combined response for the main dashboard endpoint
type DashboardResponse struct {
	EmployeeSummary       EmployeeSummaryResponse       `json:"employee_summary"`
	EmployeeCurrentNumber EmployeeCurrentNumberResponse `json:"employee_current_number"`
	EmployeeStatusStats   EmployeeStatusStatsResponse   `json:"employee_status_stats"`
	AttendanceStats       AttendanceStatsResponse       `json:"attendance_stats"`
	MonthlyAttendance     MonthlyAttendanceResponse     `json:"monthly_attendance"`
}

// ========== EMPLOYEE SUMMARY ==========

// EmployeeSummaryResponse contains total, new, active, resigned employee counts
type EmployeeSummaryResponse struct {
	TotalEmployee    int64  `json:"total_employee"`
	NewEmployee      int64  `json:"new_employee"`      // hired within 30 days
	ActiveEmployee   int64  `json:"active_employee"`   // employment_status = 'active'
	ResignedEmployee int64  `json:"resigned_employee"` // employment_status = 'resigned'
	UpdatedAt        string `json:"updated_at"`
}

// ========== EMPLOYEE CURRENT NUMBER (by month) ==========

// EmployeeCurrentNumberResponse represents employee counts by status for bar chart
type EmployeeCurrentNumberResponse struct {
	New    int64  `json:"new"`    // New hires in that month
	Active int64  `json:"active"` // Active employees in that month
	Resign int64  `json:"resign"` // Resigned employees in that month
	Month  string `json:"month"`  // Format: "YYYY-MM"
}

// ========== EMPLOYEE STATUS STATS (by employment type) ==========

// EmployeeStatusStatsResponse represents employee distribution by employment type
type EmployeeStatusStatsResponse struct {
	Permanent  int64  `json:"permanent"`
	Probation  int64  `json:"probation"`
	Contract   int64  `json:"contract"`
	Internship int64  `json:"internship"`
	Freelance  int64  `json:"freelance"`
	Month      string `json:"month"` // Format: "YYYY-MM"
}

// ========== DAILY ATTENDANCE STATS (pie chart) ==========

// AttendanceStatsResponse represents attendance statistics for a specific day
type AttendanceStatsResponse struct {
	OnTime        int64   `json:"on_time"`
	Late          int64   `json:"late"`
	Absent        int64   `json:"absent"`
	Total         int64   `json:"total"`
	OnTimePercent float64 `json:"on_time_percent"`
	LatePercent   float64 `json:"late_percent"`
	AbsentPercent float64 `json:"absent_percent"`
	Date          string  `json:"date"` // Format: "YYYY-MM-DD"
}

// ========== MONTHLY ATTENDANCE ==========

// MonthlyAttendanceResponse represents monthly attendance summary with latest records
type MonthlyAttendanceResponse struct {
	OnTime  int64                  `json:"on_time"`
	Late    int64                  `json:"late"`
	Absent  int64                  `json:"absent"`
	Records []AttendanceRecordItem `json:"records"` // Latest 10 records
	Month   string                 `json:"month"`   // Format: "YYYY-MM"
}

// AttendanceRecordItem represents a single attendance record in the list
type AttendanceRecordItem struct {
	No           int     `json:"no"`
	EmployeeName string  `json:"employee_name"`
	Status       string  `json:"status"`
	CheckIn      *string `json:"check_in,omitempty"` // Format: "HH:MM"
}
