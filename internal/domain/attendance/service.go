package attendance

import (
	"context"
)

// AttendanceService defines business logic for attendance operations
type AttendanceService interface {
	// ClockIn processes employee check-in with full validation
	ClockIn(ctx context.Context, req ClockInRequest) (AttendanceResponse, error)

	// ClockOut processes employee check-out
	ClockOut(ctx context.Context, req ClockOutRequest) (AttendanceResponse, error)

	// GetMyAttendance retrieves attendance records for authenticated employee
	GetMyAttendance(ctx context.Context, filter MyAttendanceFilter) (ListAttendanceResponse, error)

	// ListAttendance retrieves attendance records with filters (admin/manager)
	ListAttendance(ctx context.Context, filter AttendanceFilter) (ListAttendanceResponse, error)
}
