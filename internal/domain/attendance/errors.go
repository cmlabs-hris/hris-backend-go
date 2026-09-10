package attendance

import (
	"errors"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

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
	ErrCannotRejectApproved       = errors.New("cannot reject attendance that has already been approved")
	ErrCannotApproveRejected      = errors.New("cannot approve attendance that has already been rejected")

	ErrDateFormat = validator.ValidationErrors{
		validator.ValidationError{
			Field:   "date",
			Message: "date must be in YYYY-MM-DD format",
		},
	}
	ErrClockInFormat = validator.ValidationErrors{
		validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_in_time must be in YYYY-MM-DD HH:MM:SS or HH:MM:SS format",
		},
	}
	ErrClockOutFormat = validator.ValidationErrors{
		validator.ValidationError{
			Field:   "clock_out_time",
			Message: "clock_out_time must be in YYYY-MM-DD HH:MM:SS or HH:MM:SS format",
		},
	}
)
