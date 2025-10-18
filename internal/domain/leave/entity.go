package leave

import (
	"time"
)

type LeaveType struct {
	ID          string
	CompanyID   string
	Name        string
	Description *string
}

type LeaveQuota struct {
	ID          string
	EmployeeID  string
	LeaveTypeID string
	Year        int
	TotalQuota  int
	TakenQuota  int
}

type LeaveRequestStatus string

const (
	LeaveRequestStatusWaitingApproval LeaveRequestStatus = "waiting_approval"
	LeaveRequestStatusApproved        LeaveRequestStatus = "approved"
	LeaveRequestStatusRejected        LeaveRequestStatus = "rejected"
)

type LeaveRequest struct {
	ID            string
	EmployeeID    string
	LeaveTypeID   string
	StartDate     time.Time
	EndDate       time.Time
	Reason        *string
	Status        LeaveRequestStatus
	AttachmentURL *string
	ApprovedBy    string
	CreatedAt     string
}
