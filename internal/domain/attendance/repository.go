package attendance

import (
	"context"
	"time"
)

// AttendanceRepository defines data access methods for attendance records.
// All methods include companyID parameter to prevent cross-company data access attacks.
type AttendanceRepository interface {
	// Create creates a new attendance record
	Create(ctx context.Context, attendance Attendance) (Attendance, error)

	// GetByID retrieves attendance by ID with company isolation
	GetByID(ctx context.Context, id string, companyID string) (Attendance, error)

	// GetByEmployeeAndDate retrieves attendance for specific employee on specific date
	// Used to prevent double check-in
	GetByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time, companyID string) (*Attendance, error)

	// Update updates an existing attendance record
	Update(ctx context.Context, attendance Attendance) error

	// List retrieves attendance records with filters and pagination
	List(ctx context.Context, filter AttendanceFilter, companyID string) ([]Attendance, int64, error)

	// GetMyAttendance retrieves attendance records for a specific employee
	GetMyAttendance(ctx context.Context, employeeID string, filter MyAttendanceFilter, companyID string) ([]Attendance, int64, error)

	HasCheckedInToday(ctx context.Context, employeeID string, dateLocal string, companyID string) (bool, error)
	GetOpenSession(ctx context.Context, employeeID string) (Attendance, error)
}
