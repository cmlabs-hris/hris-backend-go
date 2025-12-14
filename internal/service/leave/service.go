package leave

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/notification"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
)

type LeaveServiceImpl struct {
	db *database.DB
	leave.LeaveTypeRepository
	leave.LeaveQuotaRepository
	leave.LeaveRequestRepository
	employee.EmployeeRepository
	attendance.AttendanceRepository
	quotaService        *QuotaService
	requestService      *RequestService
	fileService         file.FileService
	notificationService notification.Service
}

// GetLeaveRequest implements leave.LeaveService.
func (l *LeaveServiceImpl) GetLeaveRequest(ctx context.Context, requestID string) (leave.LeaveRequestResponse, error) {
	// Extract claims from JWT context
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return leave.LeaveRequestResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	roleStr, _ := claims["user_role"].(string)
	role := user.Role(roleStr)

	// Fetch the leave request by ID
	request, err := l.LeaveRequestRepository.GetByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leave.LeaveRequestResponse{}, leave.ErrLeaveRequestNotFound
		}
		return leave.LeaveRequestResponse{}, fmt.Errorf("failed to get leave request: %w", err)
	}

	// If the role is employee, ensure the request belongs to the employee
	if role == user.RoleEmployee {
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			return leave.LeaveRequestResponse{}, fmt.Errorf("user_id claim is missing or invalid")
		}

		// Fetch employee by user ID
		employeeData, err := l.EmployeeRepository.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return leave.LeaveRequestResponse{}, employee.ErrEmployeeNotFound
			}
			return leave.LeaveRequestResponse{}, fmt.Errorf("failed to get employee by user ID: %w", err)
		}

		// Compare employee ID from request and context
		if request.EmployeeID != employeeData.ID {
			return leave.LeaveRequestResponse{}, leave.ErrUnauthorizedAccess
		}
	}

	// Fetch the leave type details
	leaveType, err := l.LeaveTypeRepository.GetByID(ctx, request.LeaveTypeID)
	if err != nil {
		return leave.LeaveRequestResponse{}, fmt.Errorf("failed to get leave type by ID: %w", err)
	}

	// Generate attachment URL if exists
	var attachmentURL *string
	if request.AttachmentURL != nil && *request.AttachmentURL != "" {
		fullURL, err := l.fileService.GetFileURL(ctx, *request.AttachmentURL, 0)
		if err == nil {
			attachmentURL = &fullURL
		}
	}

	// Map the request to the response
	response := leave.LeaveRequestResponse{
		ID:              request.ID,
		EmployeeID:      request.EmployeeID,
		EmployeeName:    *request.EmployeeName,
		LeaveTypeID:     request.LeaveTypeID,
		LeaveTypeName:   leaveType.Name,
		StartDate:       request.StartDate,
		EndDate:         request.EndDate,
		DurationType:    string(request.DurationType),
		TotalDays:       request.TotalDays,
		WorkingDays:     request.WorkingDays,
		Reason:          request.Reason,
		AttachmentURL:   attachmentURL,
		Status:          string(request.Status),
		SubmittedAt:     request.SubmittedAt,
		ApprovedBy:      request.ApprovedBy,
		ApprovedAt:      request.ApprovedAt,
		RejectionReason: request.RejectionReason,
	}

	return response, nil
}

// GetMyRequest implements leave.LeaveService (deprecated - use ListMyLeaveRequests).
func (l *LeaveServiceImpl) GetMyRequest(ctx context.Context, userID string, companyID string) (leave.ListLeaveRequestResponse, error) {

	leaveRequest, total, err := l.LeaveRequestRepository.GetMyRequest(ctx, userID, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leave.ListLeaveRequestResponse{}, leave.ErrLeaveRequestNotFound
		}
		return leave.ListLeaveRequestResponse{}, fmt.Errorf("failed to get leave requests: %w", err)
	}

	var leaveRequestResponses []leave.LeaveRequestResponse

	for _, request := range leaveRequest {
		leaveType, err := l.LeaveTypeRepository.GetByID(ctx, request.LeaveTypeID)
		if err != nil {
			return leave.ListLeaveRequestResponse{}, fmt.Errorf("failed to get leave type by ID: %w", err)
		}

		leaveRequestResponses = append(leaveRequestResponses, leave.LeaveRequestResponse{
			ID:            request.ID,
			EmployeeID:    request.EmployeeID,
			LeaveTypeID:   request.LeaveTypeID,
			LeaveTypeName: leaveType.Name,
			StartDate:     request.StartDate,
			EndDate:       request.EndDate,
			DurationType:  string(request.DurationType),
			TotalDays:     request.TotalDays,
			WorkingDays:   request.WorkingDays,
			Reason:        request.Reason,
			Status:        string(request.Status),
			SubmittedAt:   request.SubmittedAt,
		})
	}

	return leave.ListLeaveRequestResponse{
		TotalCount: total,
		Page:       1, // Default page value, can be adjusted based on pagination logic
		Limit:      len(leaveRequestResponses),
		Requests:   leaveRequestResponses,
	}, nil
}

// ListLeaveRequest implements leave.LeaveService.
func (l *LeaveServiceImpl) ListLeaveRequest(
	ctx context.Context,
	companyID string,
	filter leave.LeaveRequestFilter,
) (leave.ListLeaveRequestResponse, error) {

	// Validate filter
	if err := filter.Validate(); err != nil {
		return leave.ListLeaveRequestResponse{}, err
	}

	// Get requests from repository
	leaveRequests, totalCount, err := l.LeaveRequestRepository.GetByCompanyID(ctx, companyID, filter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leave.ListLeaveRequestResponse{}, leave.ErrLeaveRequestNotFound
		}
		return leave.ListLeaveRequestResponse{}, fmt.Errorf("failed to get leave requests: %w", err)
	}

	// Map to response
	var leaveRequestResponses []leave.LeaveRequestResponse
	for _, req := range leaveRequests {
		// Generate attachment URL if exists
		var attachmentURL *string
		if req.AttachmentURL != nil && *req.AttachmentURL != "" {
			fullURL, err := l.fileService.GetFileURL(ctx, *req.AttachmentURL, 0)
			if err == nil {
				attachmentURL = &fullURL
			}
		}

		leaveRequestResponses = append(leaveRequestResponses, leave.LeaveRequestResponse{
			ID:              req.ID,
			EmployeeID:      req.EmployeeID,
			EmployeeName:    *req.EmployeeName,
			LeaveTypeID:     req.LeaveTypeID,
			LeaveTypeName:   *req.LeaveTypeName,
			StartDate:       req.StartDate,
			EndDate:         req.EndDate,
			DurationType:    string(req.DurationType),
			TotalDays:       req.TotalDays,
			WorkingDays:     req.WorkingDays,
			Reason:          req.Reason,
			AttachmentURL:   attachmentURL,
			Status:          string(req.Status),
			SubmittedAt:     req.SubmittedAt,
			ApprovedBy:      req.ApprovedBy,
			ApprovedAt:      req.ApprovedAt,
			RejectionReason: req.RejectionReason,
		})
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.Limit)))

	// Calculate "showing" text
	start := (filter.Page-1)*filter.Limit + 1
	end := start + len(leaveRequestResponses) - 1
	if end > int(totalCount) {
		end = int(totalCount)
	}

	showing := fmt.Sprintf("%d-%d of %d results", start, end, totalCount)
	if totalCount == 0 {
		showing = "0 results"
	}

	return leave.ListLeaveRequestResponse{
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
		Showing:    showing,
		Requests:   leaveRequestResponses,
	}, nil
}

// ListMyLeaveRequests implements leave.LeaveService - filtered version for authenticated user
func (l *LeaveServiceImpl) ListMyLeaveRequests(
	ctx context.Context,
	employeeID string,
	companyID string,
	filter leave.MyLeaveRequestFilter,
) (leave.ListLeaveRequestResponse, error) {

	// Validate filter
	if err := filter.Validate(); err != nil {
		return leave.ListLeaveRequestResponse{}, err
	}

	// Get requests from repository
	leaveRequests, totalCount, err := l.LeaveRequestRepository.GetMyRequests(ctx, employeeID, companyID, filter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leave.ListLeaveRequestResponse{}, leave.ErrLeaveRequestNotFound
		}
		return leave.ListLeaveRequestResponse{}, fmt.Errorf("failed to get my leave requests: %w", err)
	}

	// Map to response
	var leaveRequestResponses []leave.LeaveRequestResponse
	for _, req := range leaveRequests {
		// Generate attachment URL if exists
		var attachmentURL *string
		if req.AttachmentURL != nil && *req.AttachmentURL != "" {
			fullURL, err := l.fileService.GetFileURL(ctx, *req.AttachmentURL, 0)
			if err == nil {
				attachmentURL = &fullURL
			}
		}

		leaveRequestResponses = append(leaveRequestResponses, leave.LeaveRequestResponse{
			ID:              req.ID,
			EmployeeID:      req.EmployeeID,
			EmployeeName:    *req.EmployeeName,
			LeaveTypeID:     req.LeaveTypeID,
			LeaveTypeName:   *req.LeaveTypeName,
			StartDate:       req.StartDate,
			EndDate:         req.EndDate,
			DurationType:    string(req.DurationType),
			TotalDays:       req.TotalDays,
			WorkingDays:     req.WorkingDays,
			Reason:          req.Reason,
			AttachmentURL:   attachmentURL,
			Status:          string(req.Status),
			SubmittedAt:     req.SubmittedAt,
			ApprovedBy:      req.ApprovedBy,
			ApprovedAt:      req.ApprovedAt,
			RejectionReason: req.RejectionReason,
		})
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.Limit)))

	// Calculate "showing" text
	start := (filter.Page-1)*filter.Limit + 1
	end := start + len(leaveRequestResponses) - 1
	if end > int(totalCount) {
		end = int(totalCount)
	}

	showing := fmt.Sprintf("%d-%d of %d results", start, end, totalCount)
	if totalCount == 0 {
		showing = "0 results"
	}

	return leave.ListLeaveRequestResponse{
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
		Showing:    showing,
		Requests:   leaveRequestResponses,
	}, nil
}

// CreateLeaveRequest implements leave.LeaveService.
func (l *LeaveServiceImpl) GetMyQuota(ctx context.Context, userID string, year int) ([]leave.LeaveQuotaResponse, error) {
	var leaveQuotaReponse []leave.LeaveQuotaResponse

	emp, err := l.EmployeeRepository.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, employee.ErrEmployeeNotFound
		} else {
			return nil, fmt.Errorf("failed to get employee by ID: %w", err)
		}
	}

	leaveQuotas, err := l.LeaveQuotaRepository.GetByEmployeeYear(ctx, emp.ID, year)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, leave.ErrQuotaNotFound
		}
		return nil, fmt.Errorf("failed to get leave quotas: %w", err)
	}

	for _, leaveQuota := range leaveQuotas {
		leaveType, err := l.LeaveTypeRepository.GetByID(ctx, leaveQuota.LeaveTypeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get leave type by ID: %w", err)
		}
		leaveQuotaReponse = append(leaveQuotaReponse, leave.LeaveQuotaResponse{
			ID:              leaveQuota.ID,
			EmployeeID:      leaveQuota.EmployeeID,
			LeaveTypeID:     leaveQuota.LeaveTypeID,
			LeaveTypeName:   leaveType.Name,
			Year:            leaveQuota.Year,
			OpeningBalance:  *leaveQuota.OpeningBalance,
			EarnedQuota:     *leaveQuota.EarnedQuota,
			RolloverQuota:   *leaveQuota.RolloverQuota,
			AdjustmentQuota: *leaveQuota.AdjustmentQuota,
			UsedQuota:       *leaveQuota.UsedQuota,
			PendingQuota:    *leaveQuota.PendingQuota,
			AvailableQuota:  *leaveQuota.AvailableQuota,
		})
	}

	return leaveQuotaReponse, nil
}

// AdjustLeaveQuota implements leave.LeaveService.
func (l *LeaveServiceImpl) AdjustLeaveQuota(ctx context.Context, req leave.AdjustQuotaRequest) error {
	return postgresql.WithTransaction(ctx, l.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		return l.quotaService.AdjustQuota(txCtx, req.EmployeeID, req.LeaveTypeID, req.Year, req.Adjustment, req.Reason)
	})
}

// GetByEmployeeTypeYear implements leave.LeaveService.
// Subtle: this method shadows the method (LeaveQuotaRepository).GetByEmployeeTypeYear of LeaveServiceImpl.LeaveQuotaRepository.
func (l *LeaveServiceImpl) GetByEmployeeTypeYear(ctx context.Context, employeeID string, leaveTypeID string, year int) (leave.LeaveQuota, error) {
	panic("unimplemented")
}

// GetByEmployeeYear implements leave.LeaveService.
// Subtle: this method shadows the method (LeaveQuotaRepository).GetByEmployeeYear of LeaveServiceImpl.LeaveQuotaRepository.
func (l *LeaveServiceImpl) GetByEmployeeYear(ctx context.Context, employeeID string, year int) ([]leave.LeaveQuota, error) {
	panic("unimplemented")
}

// ApproveLeaveRequest implements leave.LeaveService.
func (l *LeaveServiceImpl) ApproveLeaveRequest(ctx context.Context, requestID string) error {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	approverID, ok := claims["user_id"].(string)
	if !ok || approverID == "" {
		return fmt.Errorf("user_id claim is missing or invalid")
	}

	companyID, _ := claims["company_id"].(string)

	var request leave.LeaveRequest
	err = postgresql.WithTransaction(ctx, l.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		var txErr error
		request, txErr = l.requestService.Approve(txCtx, requestID, approverID)
		if txErr != nil {
			return fmt.Errorf("failed to approve leave request: %w", txErr)
		}

		if txErr = l.quotaService.MovePendingToUsed(ctx, request.EmployeeID, request.LeaveTypeID, request.WorkingDays); txErr != nil {
			return fmt.Errorf("failed to move pending to used quota: %w", txErr)
		}

		// Create attendance records for each day of the leave period
		if txErr = l.createLeaveAttendanceRecords(txCtx, request, companyID, approverID); txErr != nil {
			return fmt.Errorf("failed to create leave attendance records: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Notify employee that their leave request was approved
	go l.notifyEmployeeOnLeaveApproved(ctx, request, companyID, approverID)

	return nil
}

// createLeaveAttendanceRecords creates attendance records with status as leave type name for each day in the leave period
func (l *LeaveServiceImpl) createLeaveAttendanceRecords(ctx context.Context, request leave.LeaveRequest, companyID string, approverID string) error {
	// Get leave type name for attendance status
	leaveType, err := l.LeaveTypeRepository.GetByID(ctx, request.LeaveTypeID)
	if err != nil {
		return fmt.Errorf("failed to get leave type: %w", err)
	}

	currentDate := request.StartDate
	now := time.Now()

	for !currentDate.After(request.EndDate) {
		// Skip weekends (Saturday = 6, Sunday = 0)
		weekday := currentDate.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// Check if attendance already exists for this date
		existingAttendance, err := l.AttendanceRepository.GetByEmployeeAndDate(ctx, request.EmployeeID, currentDate, companyID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to check existing attendance: %w", err)
		}

		// Skip if attendance already exists
		if existingAttendance != nil {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// Create attendance record for leave (WorkScheduleTimeID is nil for leave records)
		leaveAttendance := attendance.Attendance{
			EmployeeID:  request.EmployeeID,
			Date:        currentDate,
			Status:      leaveType.Name, // Use leave type name as status
			CompanyID:   companyID,
			LeaveTypeID: &request.LeaveTypeID,
			ApprovedBy:  &approverID,
			ApprovedAt:  &now,
		}

		_, err = l.AttendanceRepository.Create(ctx, leaveAttendance)
		if err != nil {
			return fmt.Errorf("failed to create leave attendance for date %s: %w", currentDate.Format("2006-01-02"), err)
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return nil
}

// CancelLeaveRequest implements leave.LeaveService.
func (l *LeaveServiceImpl) CancelLeaveRequest(ctx context.Context, requestID string) error {
	panic("unimplemented")
}

// CreateLeaveQuota implements leave.LeaveService.
func (l *LeaveServiceImpl) CreateLeaveQuota(ctx context.Context, req leave.CreateLeaveQuotaRequest) (leave.LeaveQuota, error) {
	panic("unimplemented")
}

// CreateLeaveRequest implements leave.LeaveService.
func (l *LeaveServiceImpl) CreateLeaveRequest(ctx context.Context, req leave.CreateLeaveRequestRequest) (leave.LeaveRequestResponse, error) {
	var requestResponse leave.LeaveRequestResponse
	err := postgresql.WithTransaction(ctx, l.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		leaveType, err := l.LeaveTypeRepository.GetByID(txCtx, req.LeaveTypeID)
		if err != nil {
			return fmt.Errorf("failed to get leave type by ID: %w", err)
		}
		if *leaveType.RequiresAttachment {
			if req.File == nil || req.FileHeader == nil {
				return leave.ErrAttachmentRequired
			}
			if req.FileHeader.Size > 5<<20 {
				return leave.ErrFileSizeExceeds
			}

			allowedExts := []string{".pdf", ".jpg", ".jpeg", ".png"}
			ext := strings.ToLower(filepath.Ext(req.FileHeader.Filename))

			isValidExt := false
			for _, allowed := range allowedExts {
				if ext == allowed {
					isValidExt = true
					break
				}
			}

			if !isValidExt {
				return leave.ErrFileTypeNotAllowed
			}

			attachmentURL, err := l.fileService.UploadLeaveAttachment(ctx, req.EmployeeID, req.File, req.FileHeader.Filename)
			if err != nil {
				return fmt.Errorf("failed to upload leave attachment: %w", err)
			}
			req.AttachmentURL = &attachmentURL
		}
		leaveRequest, err := l.requestService.CreateRequest(txCtx, req)
		if err != nil {
			return fmt.Errorf("failed to create leave request: %w", err)
		}

		err = l.quotaService.ReserveQuota(txCtx, leaveRequest.EmployeeID, leaveRequest.LeaveTypeID, leaveRequest.WorkingDays)
		if err != nil {
			return fmt.Errorf("failed to reserve quota: %w", err)
		}
		requestResponse = leave.LeaveRequestResponse{
			ID:            leaveRequest.ID,
			EmployeeID:    leaveRequest.EmployeeID,
			EmployeeName:  *leaveRequest.EmployeeName,
			LeaveTypeID:   leaveRequest.LeaveTypeID,
			LeaveTypeName: *leaveRequest.LeaveTypeName,
			StartDate:     leaveRequest.StartDate,
			EndDate:       leaveRequest.EndDate,
			DurationType:  string(leaveRequest.DurationType),
			TotalDays:     leaveRequest.TotalDays,
			WorkingDays:   leaveRequest.WorkingDays,
			Reason:        leaveRequest.Reason,
			AttachmentURL: leaveRequest.AttachmentURL,
			Status:        string(leaveRequest.Status),
			SubmittedAt:   leaveRequest.SubmittedAt,
		}
		return nil
	})
	if err != nil {
		return leave.LeaveRequestResponse{}, err
	}

	// Notify managers about new leave request
	go l.notifyManagersOnLeaveRequest(ctx, requestResponse)

	return requestResponse, nil
}

// CreateLeaveType implements leave.LeaveService.
func (l *LeaveServiceImpl) CreateLeaveType(ctx context.Context, req leave.CreateLeaveTypeRequest) (leave.LeaveType, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return leave.LeaveType{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return leave.LeaveType{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	_, err = l.LeaveTypeRepository.GetByName(ctx, companyID, req.Name)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return leave.LeaveType{}, fmt.Errorf("failed to get leave type by name: %w", err)
		}
	}
	if err == nil {
		return leave.LeaveType{}, leave.ErrLeaveTypeNameExists
	}

	quotaRules, err := leave.ConvertMapToQuotaRules(req.QuotaRules)
	if err != nil {
		return leave.LeaveType{}, fmt.Errorf("failed to convert quota rules: %w", err)
	}

	leaveType := leave.LeaveType{
		CompanyID:                   companyID,
		Name:                        req.Name,
		Code:                        req.Code,
		Description:                 req.Description,
		Color:                       req.Color,
		IsActive:                    req.IsActive,
		RequiresApproval:            req.RequiresApproval,
		RequiresAttachment:          req.RequiresAttachment,
		AttachmentRequiredAfterDays: req.AttachmentRequiredAfterDays,
		HasQuota:                    req.HasQuota,
		AccrualMethod:               req.AccrualMethod,
		DeductionType:               req.DeductionType,
		AllowHalfDay:                req.AllowHalfDay,
		MaxDaysPerRequest:           req.MaxDaysPerRequest,
		MinNoticeDays:               req.MinNoticeDays,
		MaxAdvanceDays:              req.MaxAdvanceDays,
		AllowBackdate:               req.AllowBackdate,
		BackdateMaxDays:             req.BackdateMaxDays,
		AllowRollover:               req.AllowRollover,
		MaxRolloverDays:             req.MaxRolloverDays,
		RolloverExpiryMonth:         req.RolloverExpiryMonth,
		QuotaCalculationType:        req.QuotaCalculationType,
		QuotaRules:                  quotaRules,
	}
	leaveType, err = l.LeaveTypeRepository.Create(ctx, leaveType)
	if err != nil {
		return leave.LeaveType{}, fmt.Errorf("failed to create leave type: %w", err)
	}

	if leaveType.HasQuota != nil && *leaveType.HasQuota {
		go func() {
			err := l.quotaService.AllocateTypeQuota(ctx, leaveType, companyID, time.Now().Year())
			if err != nil {
				fmt.Printf("failed to allocate type quota for leave type %s: %v\n", leaveType.ID, err)
			} else {
				fmt.Printf("successfully allocated type quota for leave type %s\n", leaveType.ID)
			}
		}()
	}

	return leaveType, nil
}

// DeleteLeaveQuota implements leave.LeaveService.
func (l *LeaveServiceImpl) DeleteLeaveQuota(ctx context.Context, id string) error {
	panic("unimplemented")
}

// DeleteLeaveType implements leave.LeaveService.
func (l *LeaveServiceImpl) DeleteLeaveType(ctx context.Context, id string) error {
	_, err := l.LeaveTypeRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leave.ErrLeaveTypeNotFound
		} else {
			return fmt.Errorf("failed to get leave type by id: %w", err)
		}
	} else {
		if err := l.LeaveTypeRepository.Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete leave type: %w", err)
		}
	}

	return nil
}

// GetLeaveQuota implements leave.LeaveService.
func (l *LeaveServiceImpl) GetLeaveQuota(ctx context.Context, id string) (leave.LeaveQuota, error) {
	// Extract claims from JWT context
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return leave.LeaveQuota{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	roleStr, _ := claims["user_role"].(string)
	role := user.Role(roleStr)

	leaveQuota, err := l.LeaveQuotaRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leave.LeaveQuota{}, leave.ErrQuotaNotFound
		}
		return leave.LeaveQuota{}, fmt.Errorf("failed to get leave quota: %w", err)
	}

	// If the role is employee, ensure the quota belongs to the employee
	if role == user.RoleEmployee {
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			return leave.LeaveQuota{}, fmt.Errorf("user_id claim is missing or invalid")
		}

		emp, err := l.EmployeeRepository.GetByUserID(ctx, userID)
		if err != nil {
			return leave.LeaveQuota{}, fmt.Errorf("failed to get employee by user ID: %w", err)
		}

		if leaveQuota.EmployeeID != emp.ID {
			return leave.LeaveQuota{}, leave.ErrUnauthorizedAccessQuota
		}
	}

	return leaveQuota, nil
}

// GetLeaveType implements leave.LeaveService.
func (l *LeaveServiceImpl) GetLeaveType(ctx context.Context, id string) (leave.LeaveType, error) {
	leaveType, err := l.LeaveTypeRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leave.LeaveType{}, leave.ErrLeaveTypeNotFound
		} else {
			return leave.LeaveType{}, fmt.Errorf("failed to get leave type: %w", err)
		}
	}

	return leaveType, nil
}

// ListLeaveQuota implements leave.LeaveService.
func (l *LeaveServiceImpl) ListLeaveQuota(ctx context.Context, companyID string) ([]leave.LeaveQuotaResponse, error) {
	var leaveQuotaResponse []leave.LeaveQuotaResponse

	leaveQuotas, err := l.LeaveQuotaRepository.GetByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, leave.ErrQuotaNotFound
		} else {
			return nil, fmt.Errorf("failed to get leave quota: %w", err)
		}
	}

	for _, leaveQuota := range leaveQuotas {
		leaveType, err := l.LeaveTypeRepository.GetByID(ctx, leaveQuota.LeaveTypeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get leave type by ID: %w", err)
		}
		leaveQuotaResponse = append(leaveQuotaResponse, leave.LeaveQuotaResponse{
			ID:              leaveQuota.ID,
			EmployeeID:      leaveQuota.EmployeeID,
			LeaveTypeID:     leaveQuota.LeaveTypeID,
			LeaveTypeName:   leaveType.Name,
			Year:            leaveQuota.Year,
			OpeningBalance:  *leaveQuota.OpeningBalance,
			EarnedQuota:     *leaveQuota.EarnedQuota,
			RolloverQuota:   *leaveQuota.RolloverQuota,
			AdjustmentQuota: *leaveQuota.AdjustmentQuota,
			UsedQuota:       *leaveQuota.UsedQuota,
			PendingQuota:    *leaveQuota.PendingQuota,
			AvailableQuota:  *leaveQuota.AvailableQuota,
			EmployeeName:    leaveQuota.EmployeeName,
		})
	}

	return leaveQuotaResponse, nil
}

// ListLeaveType implements leave.LeaveService.
func (l *LeaveServiceImpl) ListLeaveType(ctx context.Context, companyID string) ([]leave.LeaveTypeResponse, error) {
	leaveTypes, err := l.LeaveTypeRepository.GetByCompanyID(ctx, companyID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("failed to get leave types: %w", err)
		}
	}

	if len(leaveTypes) == 0 {
		return nil, leave.ErrLeaveTypesNotFound
	}

	var leaveTypeResponse []leave.LeaveTypeResponse
	for _, leaveType := range leaveTypes {
		leaveTypeResponse = append(leaveTypeResponse, leave.LeaveTypeResponse{
			ID:                   leaveType.ID,
			CompanyID:            leaveType.CompanyID,
			Name:                 leaveType.Name,
			Code:                 leaveType.Code,
			Description:          leaveType.Description,
			Color:                leaveType.Color,
			IsActive:             leaveType.IsActive,
			RequiresApproval:     leaveType.RequiresApproval,
			HasQuota:             leaveType.HasQuota,
			AccrualMethod:        leaveType.AccrualMethod,
			QuotaCalculationType: leaveType.QuotaCalculationType,
			QuotaRules:           leaveType.QuotaRules,
		})
	}

	return leaveTypeResponse, nil
}

// RejectLeaveRequest implements leave.LeaveService.
func (l *LeaveServiceImpl) RejectLeaveRequest(ctx context.Context, req leave.RejectRequestRequest) error {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	approverID, ok := claims["user_id"].(string)
	if !ok || approverID == "" {
		return fmt.Errorf("user_id claim is missing or invalid")
	}

	companyID, _ := claims["company_id"].(string)

	var request leave.LeaveRequest
	err = postgresql.WithTransaction(ctx, l.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		var txErr error
		request, txErr = l.requestService.Reject(ctx, req.RequestID, *req.Reason, approverID)
		if txErr != nil {
			return txErr
		}

		// Release reserved quota
		txErr = l.quotaService.ReleaseQuota(txCtx, request.EmployeeID, request.LeaveTypeID, request.WorkingDays)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Notify employee that their leave request was rejected
	go l.notifyEmployeeOnLeaveRejected(ctx, request, companyID, approverID, *req.Reason)

	return nil
}

// UpdateLeaveQuota implements leave.LeaveService.
func (l *LeaveServiceImpl) UpdateLeaveQuota(ctx context.Context, req leave.UpdateLeaveQuotaRequest) error {
	panic("unimplemented")
}

// UpdateLeaveType implements leave.LeaveService.
func (l *LeaveServiceImpl) UpdateLeaveType(ctx context.Context, req leave.UpdateLeaveTypeRequest) error {
	if err := l.LeaveTypeRepository.Update(ctx, req); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leave.ErrLeaveTypeNotFound
		}
		return fmt.Errorf("failed to update leave type: %w", err)
	}
	return nil
}

// notifyManagersOnLeaveRequest sends notifications to all managers when a leave request is submitted
func (l *LeaveServiceImpl) notifyManagersOnLeaveRequest(ctx context.Context, req leave.LeaveRequestResponse) {
	if l.notificationService == nil {
		return
	}

	// Get employee to find companyID and userID
	emp, err := l.EmployeeRepository.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return
	}

	// Get managers of the company
	managers, err := l.EmployeeRepository.GetManagersByCompanyID(ctx, emp.CompanyID)
	if err != nil {
		return
	}

	for _, manager := range managers {
		if manager.UserID == nil {
			continue
		}

		_ = l.notificationService.QueueNotification(ctx, notification.CreateNotificationRequest{
			CompanyID:   emp.CompanyID,
			RecipientID: *manager.UserID,
			SenderID:    emp.UserID,
			Type:        notification.TypeLeaveRequest,
			Title:       "New Leave Request",
			Message:     fmt.Sprintf("%s submitted a %s request from %s to %s", req.EmployeeName, req.LeaveTypeName, req.StartDate.Format("02 Jan 2006"), req.EndDate.Format("02 Jan 2006")),
			Data: map[string]interface{}{
				"employee_id":     req.EmployeeID,
				"leave_request_id": req.ID,
				"leave_type":      req.LeaveTypeName,
				"start_date":      req.StartDate.Format("2006-01-02"),
				"end_date":        req.EndDate.Format("2006-01-02"),
				"total_days":      req.TotalDays,
			},
		})
	}
}

// notifyEmployeeOnLeaveApproved sends notification to employee when leave is approved
func (l *LeaveServiceImpl) notifyEmployeeOnLeaveApproved(ctx context.Context, req leave.LeaveRequest, companyID, approverID string) {
	if l.notificationService == nil {
		return
	}

	// Get employee user ID
	emp, err := l.EmployeeRepository.GetByID(ctx, req.EmployeeID)
	if err != nil || emp.UserID == nil {
		return
	}

	// Get leave type name
	leaveType, err := l.LeaveTypeRepository.GetByID(ctx, req.LeaveTypeID)
	if err != nil {
		return
	}

	_ = l.notificationService.QueueNotification(ctx, notification.CreateNotificationRequest{
		CompanyID:   companyID,
		RecipientID: *emp.UserID,
		SenderID:    &approverID,
		Type:        notification.TypeLeaveApproved,
		Title:       "Leave Request Approved",
		Message:     fmt.Sprintf("Your %s request from %s to %s has been approved", leaveType.Name, req.StartDate.Format("02 Jan 2006"), req.EndDate.Format("02 Jan 2006")),
		Data: map[string]interface{}{
			"leave_request_id": req.ID,
			"leave_type":       leaveType.Name,
			"start_date":       req.StartDate.Format("2006-01-02"),
			"end_date":         req.EndDate.Format("2006-01-02"),
		},
	})
}

// notifyEmployeeOnLeaveRejected sends notification to employee when leave is rejected
func (l *LeaveServiceImpl) notifyEmployeeOnLeaveRejected(ctx context.Context, req leave.LeaveRequest, companyID, approverID, reason string) {
	if l.notificationService == nil {
		return
	}

	// Get employee user ID
	emp, err := l.EmployeeRepository.GetByID(ctx, req.EmployeeID)
	if err != nil || emp.UserID == nil {
		return
	}

	// Get leave type name
	leaveType, err := l.LeaveTypeRepository.GetByID(ctx, req.LeaveTypeID)
	if err != nil {
		return
	}

	_ = l.notificationService.QueueNotification(ctx, notification.CreateNotificationRequest{
		CompanyID:   companyID,
		RecipientID: *emp.UserID,
		SenderID:    &approverID,
		Type:        notification.TypeLeaveRejected,
		Title:       "Leave Request Rejected",
		Message:     fmt.Sprintf("Your %s request from %s to %s has been rejected. Reason: %s", leaveType.Name, req.StartDate.Format("02 Jan 2006"), req.EndDate.Format("02 Jan 2006"), reason),
		Data: map[string]interface{}{
			"leave_request_id": req.ID,
			"leave_type":       leaveType.Name,
			"start_date":       req.StartDate.Format("2006-01-02"),
			"end_date":         req.EndDate.Format("2006-01-02"),
			"reason":           reason,
		},
	})
}

func NewLeaveService(
	db *database.DB,
	leaveTypeRepo leave.LeaveTypeRepository,
	leaveQuotaRepo leave.LeaveQuotaRepository,
	leaveRequestRepo leave.LeaveRequestRepository,
	employeeRepo employee.EmployeeRepository,
	attendanceRepo attendance.AttendanceRepository,
	quotaService *QuotaService,
	requestService *RequestService,
	fileService file.FileService,
	notificationService notification.Service,
) leave.LeaveService {
	return &LeaveServiceImpl{
		db:                     db,
		LeaveTypeRepository:    leaveTypeRepo,
		LeaveQuotaRepository:   leaveQuotaRepo,
		LeaveRequestRepository: leaveRequestRepo,
		EmployeeRepository:     employeeRepo,
		AttendanceRepository:   attendanceRepo,
		quotaService:           quotaService,
		requestService:         requestService,
		fileService:            fileService,
		notificationService:    notificationService,
	}
}
