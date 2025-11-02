package leave

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	"github.com/jackc/pgx/v5"
)

type LeaveServiceImpl struct {
	db *database.DB
	leave.LeaveTypeRepository
	leave.LeaveQuotaRepository
	leave.LeaveRequestRepository
	employee.EmployeeRepository
	quotaService   *QuotaService
	requestService *RequestService
	fileService    file.FileService
}

// GetLeaveRequest implements leave.LeaveService.
func (l *LeaveServiceImpl) GetLeaveRequest(ctx context.Context, requestID string) (leave.LeaveRequestResponse, error) {
	// Get role from context
	role, _ := ctx.Value("user_role").(user.Role)

	// Fetch the leave request by ID
	request, err := l.LeaveRequestRepository.GetByID(ctx, requestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return leave.LeaveRequestResponse{}, leave.ErrLeaveRequestNotFound
		}
		return leave.LeaveRequestResponse{}, fmt.Errorf("failed to get leave request: %w", err)
	}

	// If the role is employee, ensure the request belongs to the employee
	if role == user.RoleEmployee {
		userID := ctx.Value("user_id").(string)

		// Fetch employee by user ID
		employee, err := l.EmployeeRepository.GetByUserID(ctx, userID)
		if err != nil {
			return leave.LeaveRequestResponse{}, fmt.Errorf("failed to get employee by user ID: %w", err)
		}

		// Compare employee ID from request and context
		if request.EmployeeID != employee.ID {
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
		if err == pgx.ErrNoRows {
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
		if err == pgx.ErrNoRows {
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
		if err == pgx.ErrNoRows {
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
		if err == pgx.ErrNoRows {
			return nil, employee.ErrEmployeeNotFound
		} else {
			return nil, fmt.Errorf("failed to get employee by ID: %w", err)
		}
	}

	leaveQuotas, err := l.LeaveQuotaRepository.GetByEmployeeYear(ctx, emp.ID, year)
	if err != nil {
		if err == pgx.ErrNoRows {
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
	approverID := ctx.Value("user_id").(string)

	return postgresql.WithTransaction(ctx, l.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		request, err := l.requestService.Approve(txCtx, requestID, approverID)
		if err != nil {
			return fmt.Errorf("failed to approve leave request: %w", err)
		}

		if err := l.quotaService.MovePendingToUsed(ctx, request.EmployeeID, request.LeaveTypeID, request.WorkingDays); err != nil {
			return fmt.Errorf("failed to move pending to used quota: %w", err)
		}

		return nil
	})
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

	return requestResponse, nil
}

// CreateLeaveType implements leave.LeaveService.
func (l *LeaveServiceImpl) CreateLeaveType(ctx context.Context, req leave.CreateLeaveTypeRequest) (leave.LeaveType, error) {
	companyID := ctx.Value("company_id").(string)
	_, err := l.LeaveTypeRepository.GetByName(ctx, companyID, req.Name)
	if err != nil {
		if err != pgx.ErrNoRows {
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
		if err == pgx.ErrNoRows {
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
	// Get role from context
	role, _ := ctx.Value("user_role").(user.Role)

	leaveQuota, err := l.LeaveQuotaRepository.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return leave.LeaveQuota{}, leave.ErrQuotaNotFound
		}
		return leave.LeaveQuota{}, fmt.Errorf("failed to get leave quota: %w", err)
	}

	// If the role is employee, ensure the quota belongs to the employee
	if role == user.RoleEmployee {
		userID := ctx.Value("user_id").(string)

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
		if err == pgx.ErrNoRows {
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
		if err == pgx.ErrNoRows {
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
		})
	}

	return leaveQuotaResponse, nil
}

// ListLeaveType implements leave.LeaveService.
func (l *LeaveServiceImpl) ListLeaveType(ctx context.Context, companyID string) ([]leave.LeaveTypeResponse, error) {
	leaveTypes, err := l.LeaveTypeRepository.GetByCompanyID(ctx, companyID)
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get leave types: %w", err)
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
	approverID := ctx.Value("user_id").(string)

	return postgresql.WithTransaction(ctx, l.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		request, err := l.requestService.Reject(ctx, req.RequestID, *req.Reason, approverID)
		if err != nil {
			return err
		}

		// Release reserved quota
		err = l.quotaService.ReleaseQuota(txCtx, request.EmployeeID, request.LeaveTypeID, request.WorkingDays)
		if err != nil {
			return err
		}
		return nil
	})
}

// UpdateLeaveQuota implements leave.LeaveService.
func (l *LeaveServiceImpl) UpdateLeaveQuota(ctx context.Context, req leave.UpdateLeaveQuotaRequest) error {
	panic("unimplemented")
}

// UpdateLeaveType implements leave.LeaveService.
func (l *LeaveServiceImpl) UpdateLeaveType(ctx context.Context, req leave.UpdateLeaveTypeRequest) error {
	if err := l.LeaveTypeRepository.Update(ctx, req); err != nil {
		return fmt.Errorf("failed to update leave type: %w", err)
	}
	return nil
}

func NewLeaveService(
	db *database.DB,
	leaveTypeRepo leave.LeaveTypeRepository,
	leaveQuotaRepo leave.LeaveQuotaRepository,
	leaveRequestRepo leave.LeaveRequestRepository,
	employeeRepo employee.EmployeeRepository,
	quotaService *QuotaService,
	requestService *RequestService,
	fileService file.FileService,
) leave.LeaveService {
	return &LeaveServiceImpl{
		db:                     db,
		LeaveTypeRepository:    leaveTypeRepo,
		LeaveQuotaRepository:   leaveQuotaRepo,
		LeaveRequestRepository: leaveRequestRepo,
		quotaService:           quotaService,
		requestService:         requestService,
		fileService:            fileService,
	}
}
