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

	// UpdateAttendance updates an attendance record (admin/manager) - for fixing wrong data
	UpdateAttendance(ctx context.Context, req UpdateAttendanceRequest) (AttendanceResponse, error)

	// GetAttendance retrieves a single attendance record by ID
	GetAttendance(ctx context.Context, id string) (AttendanceResponse, error)

	// ApproveAttendance approves an attendance record
	ApproveAttendance(ctx context.Context, req ApproveAttendanceRequest) (AttendanceResponse, error)

	// RejectAttendance rejects an attendance record with reason
	RejectAttendance(ctx context.Context, req RejectAttendanceRequest) (AttendanceResponse, error)

	// DeleteAttendance soft deletes an attendance record
	DeleteAttendance(ctx context.Context, id string) error
}
