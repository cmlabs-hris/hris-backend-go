package employee

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/invitation"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	leaveservice "github.com/cmlabs-hris/hris-backend-go/internal/service/leave"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
)

type EmployeeServiceImpl struct {
	db                *database.DB
	employeeRepo      employee.EmployeeRepository
	companyRepo       company.CompanyRepository
	fileService       file.FileService
	invitationService invitation.InvitationService
	quotaService      *leaveservice.QuotaService
}

func NewEmployeeService(
	db *database.DB,
	employeeRepo employee.EmployeeRepository,
	companyRepo company.CompanyRepository,
	fileService file.FileService,
	invitationService invitation.InvitationService,
	quotaService *leaveservice.QuotaService,
) employee.EmployeeService {
	return &EmployeeServiceImpl{
		db:                db,
		employeeRepo:      employeeRepo,
		companyRepo:       companyRepo,
		fileService:       fileService,
		invitationService: invitationService,
		quotaService:      quotaService,
	}
}

// Helper function to extract claims from context
func getClaimsFromContext(ctx context.Context) (companyID, employeeID, role string, err error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return "", "", "", fmt.Errorf("company_id claim is missing or invalid")
	}

	employeeID, _ = claims["employee_id"].(string)
	role, _ = claims["role"].(string)

	return companyID, employeeID, role, nil
}

// Helper function to map EmployeeWithDetails to EmployeeResponse
func mapEmployeeToResponse(emp employee.EmployeeWithDetails) employee.EmployeeResponse {
	var dobStr *string
	if emp.DOB != nil {
		s := emp.DOB.Format("2006-01-02")
		dobStr = &s
	}

	var resignationDateStr *string
	if emp.ResignationDate != nil {
		s := emp.ResignationDate.Format("2006-01-02")
		resignationDateStr = &s
	}

	var warningLetterStr *string
	if emp.WarningLetter != nil {
		s := string(*emp.WarningLetter)
		warningLetterStr = &s
	}

	var userID *string = emp.UserID

	var workScheduleID *string
	if emp.WorkScheduleID != "" {
		workScheduleID = &emp.WorkScheduleID
	}

	var positionID *string
	if emp.PositionID != "" {
		positionID = &emp.PositionID
	}

	var gradeID *string
	if emp.GradeID != "" {
		gradeID = &emp.GradeID
	}

	var branchID *string
	if emp.BranchID != "" {
		branchID = &emp.BranchID
	}

	var nik *string
	if emp.NIK != "" {
		nik = &emp.NIK
	}

	return employee.EmployeeResponse{
		ID:                    emp.ID,
		UserID:                userID,
		CompanyID:             emp.CompanyID,
		WorkScheduleID:        workScheduleID,
		WorkScheduleName:      emp.WorkScheduleName,
		PositionID:            positionID,
		PositionName:          emp.PositionName,
		GradeID:               gradeID,
		GradeName:             emp.GradeName,
		BranchID:              branchID,
		BranchName:            emp.BranchName,
		EmployeeCode:          emp.EmployeeCode,
		FullName:              emp.FullName,
		NIK:                   nik,
		Gender:                string(emp.Gender),
		PhoneNumber:           emp.PhoneNumber,
		Address:               emp.Address,
		PlaceOfBirth:          emp.PlaceOfBirth,
		DOB:                   dobStr,
		AvatarURL:             emp.AvatarURL,
		Education:             emp.Education,
		HireDate:              emp.HireDate.Format("2006-01-02"),
		ResignationDate:       resignationDateStr,
		EmploymentType:        string(emp.EmploymentType),
		EmploymentStatus:      string(emp.EmploymentStatus),
		WarningLetter:         warningLetterStr,
		BankName:              &emp.BankName,
		BankAccountHolderName: emp.BankAccountHolderName,
		BankAccountNumber:     &emp.BankAccountNumber,
		CreatedAt:             emp.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             emp.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// SearchEmployees implements employee.EmployeeService.
func (s *EmployeeServiceImpl) SearchEmployees(ctx context.Context, req employee.SearchEmployeeRequest) ([]employee.SearchEmployeeResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	companyID, _, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	employees, err := s.employeeRepo.Search(ctx, req.Query, companyID, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search employees: %w", err)
	}

	results := make([]employee.SearchEmployeeResponse, 0, len(employees))
	for _, emp := range employees {
		results = append(results, employee.SearchEmployeeResponse{
			ID:           emp.ID,
			EmployeeCode: emp.EmployeeCode,
			FullName:     emp.FullName,
			PositionName: emp.PositionName,
			AvatarURL:    emp.AvatarURL,
		})
	}

	return results, nil
}

// GetEmployee implements employee.EmployeeService.
func (s *EmployeeServiceImpl) GetEmployee(ctx context.Context, id string) (employee.EmployeeResponse, error) {
	companyID, requestingEmployeeID, role, err := getClaimsFromContext(ctx)
	if err != nil {
		return employee.EmployeeResponse{}, err
	}

	// Role-based access control: employees can only view their own data
	if role == "employee" && requestingEmployeeID != id {
		return employee.EmployeeResponse{}, employee.ErrUnauthorized
	}

	emp, err := s.employeeRepo.GetByIDWithDetails(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, employee.ErrEmployeeNotFound) {
			return employee.EmployeeResponse{}, employee.ErrEmployeeNotFound
		}
		return employee.EmployeeResponse{}, fmt.Errorf("failed to get employee: %w", err)
	}

	return mapEmployeeToResponse(emp), nil
}

// CreateEmployee implements employee.EmployeeService.
func (s *EmployeeServiceImpl) CreateEmployee(ctx context.Context, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
	if err := req.Validate(); err != nil {
		return employee.EmployeeResponse{}, err
	}

	companyID, inviterEmployeeID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return employee.EmployeeResponse{}, err
	}

	// Check if employee code already exists
	exists, err := s.employeeRepo.ExistsByIDOrCodeOrNIK(ctx, companyID, nil, &req.EmployeeCode, nil)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to check employee code existence: %w", err)
	}
	if exists {
		return employee.EmployeeResponse{}, employee.ErrEmployeeCodeExists
	}

	// Check if NIK already exists (if provided)
	if req.NIK != nil && *req.NIK != "" {
		exists, err = s.employeeRepo.ExistsByIDOrCodeOrNIK(ctx, companyID, nil, nil, req.NIK)
		if err != nil {
			return employee.EmployeeResponse{}, fmt.Errorf("failed to check NIK existence: %w", err)
		}
		if exists {
			return employee.EmployeeResponse{}, employee.ErrNIKExists
		}
	}

	// Check if email already has pending invitation
	hasPending, err := s.invitationService.ExistsPendingByEmail(ctx, req.Email, companyID)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to check pending invitation: %w", err)
	}
	if hasPending {
		return employee.EmployeeResponse{}, invitation.ErrEmailAlreadyInvited
	}

	// Parse dates
	hireDate, _ := time.Parse("2006-01-02", req.HireDate)

	var dob *time.Time
	if req.DOB != nil && *req.DOB != "" {
		parsed, _ := time.Parse("2006-01-02", *req.DOB)
		dob = &parsed
	}

	// Upload avatar if provided
	var avatarURL *string
	if req.File != nil && req.FileHeader != nil {
		url, err := s.fileService.UploadAvatar(ctx, req.EmployeeCode, req.File, req.FileHeader.Filename)
		if err != nil {
			return employee.EmployeeResponse{}, fmt.Errorf("failed to upload avatar: %w", err)
		}
		avatarURL = &url
	}

	// Prepare warning letter
	var warningLetter *employee.WarningLetter
	if req.WarningLetter != nil && *req.WarningLetter != "" {
		wl := employee.WarningLetter(*req.WarningLetter)
		warningLetter = &wl
	}

	// Get optional string values with empty string handling
	var workScheduleID, gradeID, branchID string
	if req.WorkScheduleID != nil {
		workScheduleID = *req.WorkScheduleID
	}
	if req.GradeID != nil {
		gradeID = *req.GradeID
	}
	if req.BranchID != nil {
		branchID = *req.BranchID
	}

	var nik string
	if req.NIK != nil {
		nik = *req.NIK
	}

	var bankName, bankAccountNumber string
	if req.BankName != nil {
		bankName = *req.BankName
	}
	if req.BankAccountNumber != nil {
		bankAccountNumber = *req.BankAccountNumber
	}

	newEmployee := employee.Employee{
		CompanyID:             companyID,
		WorkScheduleID:        workScheduleID,
		PositionID:            req.PositionID,
		GradeID:               gradeID,
		BranchID:              branchID,
		EmployeeCode:          req.EmployeeCode,
		FullName:              req.FullName,
		NIK:                   nik,
		Gender:                employee.Gender(req.Gender),
		PhoneNumber:           req.PhoneNumber,
		Address:               req.Address,
		PlaceOfBirth:          req.PlaceOfBirth,
		DOB:                   dob,
		AvatarURL:             avatarURL,
		Education:             req.Education,
		HireDate:              hireDate,
		EmploymentType:        employee.EmploymentType(strings.ToLower(req.EmploymentType)),
		EmploymentStatus:      employee.EmploymentStatusActive,
		WarningLetter:         warningLetter,
		BankName:              bankName,
		BankAccountHolderName: req.BankAccountHolderName,
		BankAccountNumber:     bankAccountNumber,
	}

	var createdEmployee employee.Employee
	var createdInvitation invitation.Invitation

	// Wrap employee creation and invitation in a transaction
	err = postgresql.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		// Create employee
		created, err := s.employeeRepo.Create(txCtx, newEmployee)
		if err != nil {
			return fmt.Errorf("failed to create employee: %w", err)
		}
		createdEmployee = created

		// Get company details for email template
		comp, err := s.companyRepo.GetByID(txCtx, companyID)
		if err != nil {
			return fmt.Errorf("failed to get company: %w", err)
		}

		// Get inviter employee name
		var inviterName string
		if inviterEmployeeID != "" {
			inviter, err := s.employeeRepo.GetByID(txCtx, inviterEmployeeID)
			if err == nil {
				inviterName = inviter.FullName
			}
		}
		if inviterName == "" {
			inviterName = "Admin"
		}

		// Get position name for email template
		emp, err := s.employeeRepo.GetByIDWithDetails(txCtx, created.ID, companyID)
		if err != nil {
			return fmt.Errorf("failed to get employee details: %w", err)
		}

		// Create and send invitation
		invReq := invitation.CreateRequest{
			EmployeeID:          created.ID,
			CompanyID:           companyID,
			InvitedByEmployeeID: inviterEmployeeID,
			Email:               req.Email,
			Role:                req.Role, // Pass role from employee request
			EmployeeName:        req.FullName,
			InviterName:         inviterName,
			CompanyName:         comp.Name,
			PositionName:        emp.PositionName,
		}

		fmt.Println("EmployeeID:", invReq.EmployeeID)
		fmt.Println("CompanyID:", invReq.CompanyID)
		fmt.Println("InvitedByEmployeeID:", invReq.InvitedByEmployeeID)
		fmt.Println("Email:", invReq.Email)
		fmt.Println("Role:", invReq.Role)
		fmt.Println("EmployeeName:", invReq.EmployeeName)
		fmt.Println("InviterName:", invReq.InviterName)
		fmt.Println("CompanyName:", invReq.CompanyName)
		fmt.Println("PositionName:", invReq.PositionName)
		inv, err := s.invitationService.CreateAndSend(txCtx, invReq)
		if err != nil {
			return fmt.Errorf("failed to create invitation: %w", err)
		}
		createdInvitation = inv

		// Assign leave quotas for the employee based on eligible leave types
		assignedQuotas, err := s.quotaService.AssignLeaveQuotasForEmployee(txCtx, createdEmployee, time.Now().Year())
		if err != nil {
			slog.Warn("Failed to assign leave quotas for employee", "employee_id", createdEmployee.ID, "error", err)
			// Don't fail the transaction, just log the warning
		} else {
			slog.Info("Assigned leave quotas for employee", "employee_id", createdEmployee.ID, "quota_count", len(assignedQuotas))
		}

		return nil
	})

	if err != nil {
		return employee.EmployeeResponse{}, err
	}

	// Log the invitation creation (optional)
	_ = createdInvitation // Invitation created successfully

	// Get the full details with JOINs
	emp, err := s.employeeRepo.GetByIDWithDetails(ctx, createdEmployee.ID, companyID)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to get created employee: %w", err)
	}

	return mapEmployeeToResponse(emp), nil
}

// UpdateEmployee implements employee.EmployeeService.
func (s *EmployeeServiceImpl) UpdateEmployee(ctx context.Context, req employee.UpdateEmployeeRequest) (employee.EmployeeResponse, error) {
	if err := req.Validate(); err != nil {
		return employee.EmployeeResponse{}, err
	}

	companyID, requestingEmployeeID, role, err := getClaimsFromContext(ctx)
	if err != nil {
		return employee.EmployeeResponse{}, err
	}

	// Role-based access control: employees can only update their own data
	if role == "employee" && requestingEmployeeID != req.ID {
		return employee.EmployeeResponse{}, employee.ErrUnauthorized
	}

	// Check if employee exists
	existingEmp, err := s.employeeRepo.GetByIDWithDetails(ctx, req.ID, companyID)
	if err != nil {
		if errors.Is(err, employee.ErrEmployeeNotFound) || errors.Is(err, pgx.ErrNoRows) {
			return employee.EmployeeResponse{}, employee.ErrEmployeeNotFound
		}
		return employee.EmployeeResponse{}, fmt.Errorf("failed to get employee: %w", err)
	}

	// Check for duplicate employee code if being updated
	if req.EmployeeCode != nil && *req.EmployeeCode != "" && *req.EmployeeCode != existingEmp.EmployeeCode {
		exists, err := s.employeeRepo.ExistsByIDOrCodeOrNIK(ctx, companyID, nil, req.EmployeeCode, nil)
		if err != nil {
			return employee.EmployeeResponse{}, fmt.Errorf("failed to check employee code: %w", err)
		}
		if exists {
			return employee.EmployeeResponse{}, employee.ErrEmployeeCodeExists
		}
	}

	// Check for duplicate NIK if being updated
	if req.NIK != nil && *req.NIK != "" && *req.NIK != existingEmp.NIK {
		exists, err := s.employeeRepo.ExistsByIDOrCodeOrNIK(ctx, companyID, nil, nil, req.NIK)
		if err != nil {
			return employee.EmployeeResponse{}, fmt.Errorf("failed to check NIK: %w", err)
		}
		if exists {
			return employee.EmployeeResponse{}, employee.ErrNIKExists
		}
	}

	// Perform update
	err = s.employeeRepo.Update(ctx, req.ID, companyID, req)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to update employee: %w", err)
	}

	// Get updated employee
	emp, err := s.employeeRepo.GetByIDWithDetails(ctx, req.ID, companyID)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to get updated employee: %w", err)
	}

	return mapEmployeeToResponse(emp), nil
}

// DeleteEmployee implements employee.EmployeeService.
func (s *EmployeeServiceImpl) DeleteEmployee(ctx context.Context, id string) error {
	companyID, requestingEmployeeID, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return err
	}

	// Prevent self-deletion
	if requestingEmployeeID == id {
		return employee.ErrCannotDeleteSelf
	}

	// Perform soft delete
	err = s.employeeRepo.SoftDelete(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, employee.ErrEmployeeNotFound) {
			return employee.ErrEmployeeNotFound
		}
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	return nil
}

// ListEmployees implements employee.EmployeeService.
func (s *EmployeeServiceImpl) ListEmployees(ctx context.Context, filter employee.EmployeeFilter) (employee.ListEmployeeResponse, error) {
	if err := filter.Validate(); err != nil {
		return employee.ListEmployeeResponse{}, err
	}

	companyID, _, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return employee.ListEmployeeResponse{}, err
	}

	employees, total, err := s.employeeRepo.List(ctx, filter, companyID)
	if err != nil {
		return employee.ListEmployeeResponse{}, fmt.Errorf("failed to list employees: %w", err)
	}

	responses := make([]employee.EmployeeResponse, 0, len(employees))
	for _, emp := range employees {
		responses = append(responses, mapEmployeeToResponse(emp))
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))
	showing := fmt.Sprintf("%d-%d of %d", (filter.Page-1)*filter.Limit+1, min((filter.Page)*filter.Limit, int(total)), total)
	if total == 0 {
		showing = "0 of 0"
	}

	return employee.ListEmployeeResponse{
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
		Showing:    showing,
		Employees:  responses,
	}, nil
}

// InactivateEmployee implements employee.EmployeeService.
func (s *EmployeeServiceImpl) InactivateEmployee(ctx context.Context, req employee.InactivateEmployeeRequest) (employee.EmployeeResponse, error) {
	if err := req.Validate(); err != nil {
		return employee.EmployeeResponse{}, err
	}

	companyID, _, _, err := getClaimsFromContext(ctx)
	if err != nil {
		return employee.EmployeeResponse{}, err
	}

	// Check if employee exists and is active
	existingEmp, err := s.employeeRepo.GetByIDWithDetails(ctx, req.ID, companyID)
	if err != nil {
		if errors.Is(err, employee.ErrEmployeeNotFound) {
			return employee.EmployeeResponse{}, employee.ErrEmployeeNotFound
		}
		return employee.EmployeeResponse{}, fmt.Errorf("failed to get employee: %w", err)
	}

	if existingEmp.EmploymentStatus != employee.EmploymentStatusActive {
		return employee.EmployeeResponse{}, employee.ErrEmployeeAlreadyInactive
	}

	// Inactivate employee
	err = s.employeeRepo.Inactivate(ctx, req.ID, companyID, req.ResignationDate)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to inactivate employee: %w", err)
	}

	// Get updated employee
	emp, err := s.employeeRepo.GetByIDWithDetails(ctx, req.ID, companyID)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to get updated employee: %w", err)
	}

	return mapEmployeeToResponse(emp), nil
}

// UploadAvatar implements employee.EmployeeService.
func (s *EmployeeServiceImpl) UploadAvatar(ctx context.Context, req employee.UploadAvatarRequest) (employee.EmployeeResponse, error) {
	if err := req.Validate(); err != nil {
		return employee.EmployeeResponse{}, err
	}

	companyID, requestingEmployeeID, role, err := getClaimsFromContext(ctx)
	if err != nil {
		return employee.EmployeeResponse{}, err
	}

	// Role-based access control: employees can only update their own avatar
	if role == "employee" && requestingEmployeeID != req.EmployeeID {
		return employee.EmployeeResponse{}, employee.ErrUnauthorized
	}

	// Check if employee exists
	_, err = s.employeeRepo.GetByIDWithDetails(ctx, req.EmployeeID, companyID)
	if err != nil {
		if errors.Is(err, employee.ErrEmployeeNotFound) {
			return employee.EmployeeResponse{}, employee.ErrEmployeeNotFound
		}
		return employee.EmployeeResponse{}, fmt.Errorf("failed to get employee: %w", err)
	}

	// Upload avatar
	avatarURL, err := s.fileService.UploadAvatar(ctx, req.EmployeeID, req.File, req.FileHeader.Filename)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Update employee avatar URL
	err = s.employeeRepo.UpdateAvatar(ctx, req.EmployeeID, companyID, avatarURL)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to update avatar URL: %w", err)
	}

	// Get updated employee
	emp, err := s.employeeRepo.GetByIDWithDetails(ctx, req.EmployeeID, companyID)
	if err != nil {
		return employee.EmployeeResponse{}, fmt.Errorf("failed to get updated employee: %w", err)
	}

	return mapEmployeeToResponse(emp), nil
}
