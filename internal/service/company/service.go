package company

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/branch"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/grade"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/position"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/notification"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/fixtures"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	leaveservice "github.com/cmlabs-hris/hris-backend-go/internal/service/leave"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
)

type CompanyServiceImpl struct {
	db *database.DB
	company.CompanyRepository
	fileService file.FileService
	user.UserRepository

	// Repositories for seeding default data
	positionRepo         position.PositionRepository
	gradeRepo            grade.GradeRepository
	branchRepo           branch.BranchRepository
	leaveTypeRepo        leave.LeaveTypeRepository
	workScheduleRepo     schedule.WorkScheduleRepository
	workScheduleTimeRepo schedule.WorkScheduleTimeRepository
	employeeRepo         employee.EmployeeRepository

	// Service for assigning leave quotas
	quotaService     *leaveservice.QuotaService
	notificationRepo notification.Repository
}

// UploadCompanyLogo implements company.CompanyService.
func (c *CompanyServiceImpl) UploadCompanyLogo(ctx context.Context, req company.UploadCompanyLogoRequest) (company.UploadCompanyLogoResponse, error) {
	companyData, err := c.CompanyRepository.GetByID(ctx, req.CompanyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return company.UploadCompanyLogoResponse{}, company.ErrCompanyNotFound
		}
		return company.UploadCompanyLogoResponse{}, fmt.Errorf("failed to get company by ID: %w", err)
	}

	logoURLResult, err := c.fileService.UploadCompanyLogo(ctx, companyData.Username, req.File, req.FileHeader.Filename)
	if err != nil {
		return company.UploadCompanyLogoResponse{}, fmt.Errorf("failed to upload company logo: %w", err)
	}
	if err := c.CompanyRepository.Update(ctx, req.CompanyID, company.UpdateCompanyRequest{LogoURL: &logoURLResult}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return company.UploadCompanyLogoResponse{}, company.ErrCompanyNotFound
		}
		return company.UploadCompanyLogoResponse{}, fmt.Errorf("failed to update company logo URL: %w", err)
	}

	attachmentURL, _ := c.fileService.GetFileURL(ctx, logoURLResult, 0)

	return company.UploadCompanyLogoResponse{LogoURL: attachmentURL}, nil
}

// Create implements company.CompanyService.
// Subtle: this method shadows the method (CompanyRepository).Create of CompanyServiceImpl.CompanyRepository.
func (c *CompanyServiceImpl) Create(ctx context.Context, req company.CreateCompanyRequest) (company.Company, error) {
	var newCompany company.Company
	err := postgresql.WithTransaction(ctx, c.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		_, err := c.CompanyRepository.GetByUsername(txCtx, req.Username)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("failed to get company by username: %w", err)
			}
		} else {
			return company.ErrCompanyUsernameExists
		}
		if req.File != nil && req.FileHeader != nil {
			if req.FileHeader.Size > 5<<20 {
				return company.ErrFileSizeExceeds
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

			// Use req.Username instead of newCompany.Username (which is empty at this point)
			attachmentURL, err := c.fileService.UploadCompanyLogo(txCtx, req.Username, req.File, req.FileHeader.Filename)
			if err != nil {
				return fmt.Errorf("failed to upload company logo attachment: %w", err)
			}
			attachmentURL, _ = c.fileService.GetFileURL(ctx, attachmentURL, 0)
			req.AttachmentURL = &attachmentURL
		}
		newCompany, err = c.CompanyRepository.Create(txCtx, company.Company{
			Name:     req.Name,
			Username: req.Username,
			Address:  req.Address,
			LogoURL:  req.AttachmentURL,
		})
		if err != nil {
			return fmt.Errorf("failed to create company: %w", err)
		}

		_, claims, err := jwtauth.FromContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to extract claims from context: %w", err)
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			return fmt.Errorf("user_id claim is missing or invalid")
		}
		fmt.Println("hi")
		var userIDPointer *string = &userID
		fmt.Println(userIDPointer)
		if err := c.UserRepository.Update(txCtx, user.UpdateUserRequest{ID: userID, CompanyID: &newCompany.ID}); err != nil {
			return fmt.Errorf("failed to update user with company ID: %w", err)
		}

		if err := c.UserRepository.UpdateRole(txCtx, user.UpdateUserRoleRequest{ID: userID, Role: string(user.RoleOwner)}); err != nil {
			return fmt.Errorf("failed to update user role: %w", err)
		}

		// Seed default master data for the new company
		seededIDs, err := c.seedDefaultData(txCtx, newCompany.ID, newCompany.Name)
		if err != nil {
			slog.Error("Failed to seed default data for company", "company_id", newCompany.ID, "error", err)
			return fmt.Errorf("failed to seed default data: %w", err)
		}

		// Get default IDs for the owner employee
		positionID, gradeID, branchID, workScheduleID := seededIDs.GetOwnerDefaults()

		// Create the owner as an employee (use email as temporary name, should be updated later)
		newEmployee := employee.Employee{
			UserID:            userIDPointer,
			CompanyID:         newCompany.ID,
			PositionID:        positionID,
			GradeID:           gradeID,
			BranchID:          branchID,
			WorkScheduleID:    workScheduleID,
			EmployeeCode:      "0001-0001",        // First employee code
			FullName:          "Company Owner",    // Placeholder, should be updated in onboarding
			NIK:               "0000000000000000", // Placeholder, should be updated in onboarding
			Gender:            employee.Male,      // Default, should be updated in onboarding
			PhoneNumber:       "000000000000",     // Placeholder, should be updated in onboarding
			HireDate:          time.Now(),
			EmploymentType:    employee.EmploymentTypePermanent,
			EmploymentStatus:  employee.EmploymentStatusActive,
			BankName:          "N/A", // Placeholder, should be updated in onboarding
			BankAccountNumber: "N/A", // Placeholder, should be updated in onboarding
		}

		createdOwnerEmployee, err := c.employeeRepo.Create(txCtx, newEmployee)
		if err != nil {
			return fmt.Errorf("failed to create owner employee: %w", err)
		}
		slog.Info("Created owner employee", "company_id", newCompany.ID, "user_id", userID, "employee_id", createdOwnerEmployee.ID)

		// Assign leave quotas for the owner based on eligible leave types
		assignedQuotas, err := c.quotaService.AssignLeaveQuotasForEmployee(txCtx, createdOwnerEmployee, time.Now().Year())
		if err != nil {
			slog.Warn("Failed to assign leave quotas for owner", "employee_id", createdOwnerEmployee.ID, "error", err)
			// Don't fail the transaction, just log the warning
		} else {
			slog.Info("Assigned leave quotas for owner", "employee_id", createdOwnerEmployee.ID, "quota_count", len(assignedQuotas))
		}

		// c.notificationRepo.UpsertPreference(ctx, &notification.NotificationPreference{
		// 	UserID: userID,
		// 	NotificationType: notii,
		// })

		return nil
	})
	if err != nil {
		return company.Company{}, err
	}

	return newCompany, nil
}

// seedDefaultData creates default master data for a newly created company and returns the IDs
func (c *CompanyServiceImpl) seedDefaultData(ctx context.Context, companyID string, companyName string) (*fixtures.SeededDataIDs, error) {
	seededIDs := fixtures.NewSeededDataIDs()

	// 1. Seed default positions
	positions := fixtures.GetDefaultPositions(companyID)
	for _, pos := range positions {
		createdPos, err := c.positionRepo.Create(ctx, pos)
		if err != nil {
			slog.Warn("Failed to create default position", "position", pos.Name, "error", err)
			// Continue with other positions even if one fails (might be duplicate)
		} else {
			seededIDs.PositionIDs[createdPos.Name] = createdPos.ID
		}
	}
	slog.Info("Seeded default positions", "company_id", companyID, "count", len(positions))

	// 2. Seed default grades
	grades := fixtures.GetDefaultGrades(companyID)
	for _, g := range grades {
		createdGrade, err := c.gradeRepo.Create(ctx, g)
		if err != nil {
			slog.Warn("Failed to create default grade", "grade", g.Name, "error", err)
		} else {
			seededIDs.GradeIDs[createdGrade.Name] = createdGrade.ID
		}
	}
	slog.Info("Seeded default grades", "company_id", companyID, "count", len(grades))

	// 3. Seed default branch (headquarters)
	defaultBranch := fixtures.GetDefaultBranch(companyID, companyName)
	createdBranch, err := c.branchRepo.Create(ctx, defaultBranch)
	if err != nil {
		slog.Warn("Failed to create default branch", "branch", defaultBranch.Name, "error", err)
	} else {
		seededIDs.BranchID = createdBranch.ID
		slog.Info("Seeded default branch", "company_id", companyID, "branch", defaultBranch.Name)
	}

	// 4. Seed default leave types (Indonesian labor law compliant)
	leaveTypes := fixtures.GetDefaultLeaveTypes(companyID)
	for _, lt := range leaveTypes {
		createdLT, err := c.leaveTypeRepo.Create(ctx, lt)
		if err != nil {
			slog.Warn("Failed to create default leave type", "leave_type", lt.Name, "error", err)
		} else if lt.Code != nil {
			seededIDs.LeaveTypeIDs[*lt.Code] = createdLT.ID
		}
	}
	slog.Info("Seeded default leave types", "company_id", companyID, "count", len(leaveTypes))

	// 5. Seed all default work schedules (Standard, Night Shift, Afternoon Shift, Flexible)
	allSchedules := fixtures.GetAllDefaultWorkSchedules(companyID)
	for _, scheduleDef := range allSchedules {
		createdSchedule, err := c.workScheduleRepo.Create(ctx, scheduleDef.Schedule)
		if err != nil {
			slog.Warn("Failed to create work schedule", "schedule", scheduleDef.Schedule.Name, "error", err)
			continue
		}

		// Store schedule ID
		seededIDs.WorkScheduleIDs[createdSchedule.Name] = createdSchedule.ID

		// Set the default schedule ID (Standard Office Hours) for owner employee
		if scheduleDef.Schedule.Name == "Standard Office Hours" {
			seededIDs.WorkScheduleID = createdSchedule.ID
		}

		slog.Info("Seeded work schedule", "company_id", companyID, "schedule", createdSchedule.Name)

		// 6. Seed work schedule times for this schedule
		scheduleTimes := scheduleDef.TimesGetter(createdSchedule.ID)
		seededIDs.WorkScheduleTimeIDs[createdSchedule.Name] = make(map[int]string)
		for _, st := range scheduleTimes {
			createdST, err := c.workScheduleTimeRepo.Create(ctx, st, companyID)
			if err != nil {
				slog.Warn("Failed to create schedule time", "schedule", createdSchedule.Name, "day", st.DayOfWeek, "error", err)
			} else {
				seededIDs.WorkScheduleTimeIDs[createdSchedule.Name][st.DayOfWeek] = createdST.ID
			}
		}
		slog.Info("Seeded work schedule times", "company_id", companyID, "schedule", createdSchedule.Name, "count", len(scheduleTimes))
	}

	return seededIDs, nil
}

// Delete implements company.CompanyService.
func (c *CompanyServiceImpl) Delete(ctx context.Context, id string) error {
	if err := c.CompanyRepository.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return company.ErrCompanyNotFound
		}
		return fmt.Errorf("failed to delete company with id %s: %w", id, err)
	}
	return nil
}

// GetByID implements company.CompanyService.
// Subtle: this method shadows the method (CompanyRepository).GetByID of CompanyServiceImpl.CompanyRepository.
func (c *CompanyServiceImpl) GetByID(ctx context.Context, id string) (company.CompanyResponse, error) {
	companyData, err := c.CompanyRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return company.CompanyResponse{}, company.ErrCompanyNotFound
		}
		return company.CompanyResponse{}, fmt.Errorf("failed to get company by ID: %w", err)
	}
	// Generate attachment URL if exists
	var attachmentURL *string
	if companyData.LogoURL != nil && *companyData.LogoURL != "" {
		fullURL, err := c.fileService.GetFileURL(ctx, *companyData.LogoURL, 0)
		if err == nil {
			attachmentURL = &fullURL
		}
	}
	return company.CompanyResponse{
		ID:        companyData.ID,
		Name:      companyData.Name,
		Username:  companyData.Username,
		Address:   companyData.Address,
		LogoURL:   attachmentURL,
		CreatedAt: companyData.CreatedAt,
		UpdatedAt: companyData.UpdatedAt,
	}, nil
}

// List implements company.CompanyService.
func (c *CompanyServiceImpl) List(ctx context.Context) ([]company.Company, error) {
	panic("unimplemented")
}

// Update implements company.CompanyService.
// Subtle: this method shadows the method (CompanyRepository).Update of CompanyServiceImpl.CompanyRepository.
func (c *CompanyServiceImpl) Update(ctx context.Context, id string, req company.UpdateCompanyRequest) error {
	err := c.CompanyRepository.Update(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update company with id %s: %w", id, err)
	}
	return nil
}

func NewCompanyService(
	db *database.DB,
	companyRepository company.CompanyRepository,
	fileService file.FileService,
	userRepository user.UserRepository,
	positionRepo position.PositionRepository,
	gradeRepo grade.GradeRepository,
	branchRepo branch.BranchRepository,
	leaveTypeRepo leave.LeaveTypeRepository,
	workScheduleRepo schedule.WorkScheduleRepository,
	workScheduleTimeRepo schedule.WorkScheduleTimeRepository,
	employeeRepo employee.EmployeeRepository,
	quotaService *leaveservice.QuotaService,
	notificationRepo notification.Repository,
) company.CompanyService {
	return &CompanyServiceImpl{
		db:                   db,
		CompanyRepository:    companyRepository,
		fileService:          fileService,
		UserRepository:       userRepository,
		positionRepo:         positionRepo,
		gradeRepo:            gradeRepo,
		branchRepo:           branchRepo,
		leaveTypeRepo:        leaveTypeRepo,
		workScheduleRepo:     workScheduleRepo,
		workScheduleTimeRepo: workScheduleTimeRepo,
		employeeRepo:         employeeRepo,
		quotaService:         quotaService,
		notificationRepo:     notificationRepo,
	}
}
