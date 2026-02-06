package attendance

import "errors"

// Attendance domain errors
var (
	// Check-in errors
	ErrAlreadyCheckedIn     = errors.New("you have already checked in today")
	ErrNoScheduleFound      = errors.New("no schedule found for today")
	ErrOutsideAllowedRadius = errors.New("you are outside the allowed radius")
	ErrTooEarlyToCheckIn    = errors.New("too early to check in")
	ErrNotCheckedIn         = errors.New("you have not checked in yet")
	ErrAlreadyCheckedOut    = errors.New("you have already checked out")

	// General errors
	ErrAttendanceNotFound         = errors.New("attendance record not found")
	ErrUnauthorized               = errors.New("unauthorized to access this attendance record")
	ErrAttendanceAlreadyProcessed = errors.New("attendance has already been approved or rejected")
)
