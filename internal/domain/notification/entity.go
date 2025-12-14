package notification

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	TypeAttendanceClockIn  NotificationType = "attendance_clock_in"
	TypeAttendanceClockOut NotificationType = "attendance_clock_out"
	TypeLeaveRequest       NotificationType = "leave_request"
	TypeLeaveApproved      NotificationType = "leave_approved"
	TypeLeaveRejected      NotificationType = "leave_rejected"
	TypePayrollGenerated   NotificationType = "payroll_generated"
	TypeScheduleUpdated    NotificationType = "schedule_updated"
	TypeInvitationSent     NotificationType = "invitation_sent"
	TypeEmployeeJoined     NotificationType = "employee_joined"
)

// AllNotificationTypes returns all available notification types
func AllNotificationTypes() []NotificationType {
	return []NotificationType{
		TypeAttendanceClockIn,
		TypeAttendanceClockOut,
		TypeLeaveRequest,
		TypeLeaveApproved,
		TypeLeaveRejected,
		TypePayrollGenerated,
		TypeScheduleUpdated,
		TypeInvitationSent,
		TypeEmployeeJoined,
	}
}

// Notification represents a notification entity
type Notification struct {
	ID          string
	CompanyID   string
	RecipientID string
	SenderID    *string
	Type        NotificationType
	Title       string
	Message     string
	Data        map[string]interface{}
	IsRead      bool
	ReadAt      *time.Time
	CreatedAt   time.Time
}

// NotificationPreference represents user preference for a notification type
type NotificationPreference struct {
	ID               string
	UserID           string
	NotificationType NotificationType
	EmailEnabled     bool
	PushEnabled      bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
