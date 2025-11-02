package leave

import (
	"context"
)

type LeaveService interface {
	// Type
	CreateLeaveType(ctx context.Context, req CreateLeaveTypeRequest) (LeaveType, error)
	UpdateLeaveType(ctx context.Context, req UpdateLeaveTypeRequest) error
	GetLeaveType(ctx context.Context, id string) (LeaveType, error)
	ListLeaveType(ctx context.Context, companyID string) ([]LeaveTypeResponse, error)
	DeleteLeaveType(ctx context.Context, id string) error
	// Quota
	CreateLeaveQuota(ctx context.Context, req CreateLeaveQuotaRequest) (LeaveQuota, error)
	UpdateLeaveQuota(ctx context.Context, req UpdateLeaveQuotaRequest) error
	GetLeaveQuota(ctx context.Context, id string) (LeaveQuota, error)
	GetByEmployeeTypeYear(ctx context.Context, employeeID, leaveTypeID string, year int) (LeaveQuota, error)
	GetByEmployeeYear(ctx context.Context, employeeID string, year int) ([]LeaveQuota, error)
	ListLeaveQuota(ctx context.Context, companyID string) ([]LeaveQuotaResponse, error)
	DeleteLeaveQuota(ctx context.Context, id string) error
	AdjustLeaveQuota(ctx context.Context, req AdjustQuotaRequest) error
	GetMyQuota(ctx context.Context, userID string, year int) ([]LeaveQuotaResponse, error)
	// Request
	CreateLeaveRequest(ctx context.Context, req CreateLeaveRequestRequest) (LeaveRequestResponse, error)
	ApproveLeaveRequest(ctx context.Context, requestID string) error
	RejectLeaveRequest(ctx context.Context, req RejectRequestRequest) error
	CancelLeaveRequest(ctx context.Context, requestID string) error
	ListLeaveRequest(ctx context.Context, companyID string, filter LeaveRequestFilter) (ListLeaveRequestResponse, error)
	ListMyLeaveRequests(ctx context.Context, employeeID string, companyID string, filter MyLeaveRequestFilter) (ListLeaveRequestResponse, error)
	GetMyRequest(ctx context.Context, userID string, companyID string) (ListLeaveRequestResponse, error)
	GetLeaveRequest(ctx context.Context, requestID string) (LeaveRequestResponse, error)
}
