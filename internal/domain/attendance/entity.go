package attendance

import (
	"time"
)

type Attendance struct {
	ID                 string
	EmployeeID         string
	Date               time.Time
	WorkScheduleTimeID *string
	ActualLocationType *string
	ClockIn            *time.Time
	ClockOut           *time.Time
	WorkHoursInMinutes *int
	ClockInLatitude    *float64
	ClockInLongitude   *float64
	ClockInProofURL    *string
	ClockOutLatitude   *float64
	ClockOutLongitude  *float64
	ClockOutProofURL   *string
	Status             string
	CompanyID          string
	ApprovedBy         *string
	ApprovedAt         *time.Time
	RejectionReason    *string
	LeaveTypeID        *string
	LateMinutes        *int
	EarlyLeaveMinutes  *int
	OvertimeMinutes    *int
	CreatedAt          time.Time
	UpdatedAt          time.Time

	// DTO
	EmployeeName     *string
	EmployeePosition *string
}
