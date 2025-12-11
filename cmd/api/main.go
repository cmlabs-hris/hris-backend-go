package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/config"
	appHTTP "github.com/cmlabs-hris/hris-backend-go/internal/handler/http"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/email"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/oauth"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/storage"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	attendanceService "github.com/cmlabs-hris/hris-backend-go/internal/service/attendance"
	serviceAuth "github.com/cmlabs-hris/hris-backend-go/internal/service/auth"
	serviceCompany "github.com/cmlabs-hris/hris-backend-go/internal/service/company"
	dashboardService "github.com/cmlabs-hris/hris-backend-go/internal/service/dashboard"
	employeeService "github.com/cmlabs-hris/hris-backend-go/internal/service/employee"
	employeeDashboardService "github.com/cmlabs-hris/hris-backend-go/internal/service/employee_dashboard"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	invitationService "github.com/cmlabs-hris/hris-backend-go/internal/service/invitation"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/master"
	payrollService "github.com/cmlabs-hris/hris-backend-go/internal/service/payroll"
	scheduleService "github.com/cmlabs-hris/hris-backend-go/internal/service/schedule"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	dsn := cfg.DatabaseURL()
	db, err := database.NewPostgreSQLDB(dsn)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}

	userRepo := postgresql.NewUserRepository(db)
	companyRepo := postgresql.NewCompanyRepository(db)
	JWTRepository := postgresql.NewJWTRepository(db)
	leaveTypeRepo := postgresql.NewLeaveTypeRepository(db)
	leaveQuotaRepo := postgresql.NewLeaveQuotaRepository(db)
	leaveRequestRepo := postgresql.NewLeaveRequestRepository(db)
	employeeRepo := postgresql.NewEmployeeRepository(db)
	branchRepo := postgresql.NewBranchRepository(db)
	gradeRepo := postgresql.NewGradeRepository(db)
	positionRepo := postgresql.NewPositionRepository(db)
	workScheduleRepo := postgresql.NewWorkScheduleRepository(db)
	workScheduleTimeRepo := postgresql.NewWorkScheduleTimeRepository(db)
	workScheduleLocationRepo := postgresql.NewWorkScheduleLocationRepository(db)
	employeeScheduleAssignmentRepo := postgresql.NewEmployeeScheduleAssignmentRepository(db)
	attendanceRepo := postgresql.NewAttendanceRepository(db)
	invitationRepo := postgresql.NewInvitationRepository(db)
	payrollRepo := postgresql.NewPayrollRepository(db)
	dashboardRepo := postgresql.NewDashboardRepository(db)
	empDashboardRepo := postgresql.NewEmployeeDashboardRepository(db)

	JWTService := jwt.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessExpiration, cfg.JWT.RefreshExpiration)
	GoogleService := oauth.NewGoogleService(cfg.OAuth2Google.ClientID, cfg.OAuth2Google.ClientSecret, cfg.OAuth2Google.RedirectURL, cfg.OAuth2Google.Scopes)
	quotaCalculatorService := leave.NewQuotaCalculator()
	quotaService := leave.NewQuotaService(db, leaveTypeRepo, leaveQuotaRepo, employeeRepo, quotaCalculatorService)
	requestService := leave.NewRequestService(db, leaveTypeRepo, leaveQuotaRepo, leaveRequestRepo, employeeRepo)
	var fileStorage storage.FileStorage
	switch cfg.Storage.Type {
	case "local":
		fileStorage, err = storage.NewLocalStorage(
			cfg.Storage.BasePath,
			cfg.Storage.BaseURL,
		)
		if err != nil {
			log.Fatal("Failed to initialize local storage:", err)
		}
	case "minio":
		// Future: minIO implementation
		log.Fatal("Minio storage not yet implemented")
	default:
		log.Fatal("Unsupported storage types: ", cfg.Storage.Type)
	}

	fileService := file.NewFileService(fileStorage)
	emailService, err := email.NewEmailService(cfg.SMTP)
	if err != nil {
		log.Fatal("Failed to initialize email service:", err)
	}
	authService := serviceAuth.NewAuthService(db, userRepo, companyRepo, JWTService, JWTRepository)
	companyService := serviceCompany.NewCompanyService(
		db,
		companyRepo,
		fileService,
		userRepo,
		positionRepo,
		gradeRepo,
		branchRepo,
		leaveTypeRepo,
		workScheduleRepo,
		workScheduleTimeRepo,
		employeeRepo,
		quotaService,
	)
	leaveService := leave.NewLeaveService(db, leaveTypeRepo, leaveQuotaRepo, leaveRequestRepo, employeeRepo, attendanceRepo, quotaService, requestService, fileService)
	masterService := master.NewMasterService(branchRepo, gradeRepo, positionRepo)
	scheduleService := scheduleService.NewScheduleService(
		db,
		workScheduleRepo,
		workScheduleTimeRepo,
		workScheduleLocationRepo,
		employeeScheduleAssignmentRepo,
		employeeRepo,
	)
	attendanceService := attendanceService.NewAttendanceService(
		db,
		attendanceRepo,
		employeeRepo,
		workScheduleRepo,
		workScheduleTimeRepo,
		branchRepo,
		fileService,
	)
	invitationService := invitationService.NewInvitationService(
		db,
		invitationRepo,
		employeeRepo,
		userRepo,
		emailService,
		fileService,
		cfg.Invitation,
	)
	employeeService := employeeService.NewEmployeeService(
		db,
		employeeRepo,
		companyRepo,
		fileService,
		invitationService,
		quotaService,
	)
	payrollSvc := payrollService.NewPayrollService(db, payrollRepo, employeeRepo)
	dashboardSvc := dashboardService.NewDashboardService(dashboardRepo)
	empDashboardSvc := employeeDashboardService.NewEmployeeDashboardService(empDashboardRepo)

	authHandler := appHTTP.NewAuthHandler(JWTService, authService, GoogleService, cfg.App.FrontendURL)
	companyHandler := appHTTP.NewCompanyHandler(JWTService, companyService, fileService)
	leaveHandler := appHTTP.NewLeaveHandler(leaveService, fileService)
	masterHandler := appHTTP.NewMasterHandler(masterService)
	scheduleHandler := appHTTP.NewScheduleHandler(scheduleService)
	attendanceHandler := appHTTP.NewAttendanceHandler(attendanceService)
	invitationHandler := appHTTP.NewInvitationHandler(invitationService)
	employeeHandler := appHTTP.NewEmployeeHandler(employeeService, invitationService)
	payrollHandler := appHTTP.NewPayrollHandler(payrollSvc)
	dashboardHandler := appHTTP.NewDashboardHandler(dashboardSvc)
	empDashboardHandler := appHTTP.NewEmployeeDashboardHandler(empDashboardSvc)

	router := appHTTP.NewRouter(
		JWTService,
		authHandler,
		companyHandler,
		leaveHandler,
		masterHandler,
		scheduleHandler,
		attendanceHandler,
		employeeHandler,
		invitationHandler,
		payrollHandler,
		dashboardHandler,
		empDashboardHandler,
		cfg.Storage.BasePath,
	)

	port := fmt.Sprintf(":%d", cfg.App.Port)
	fmt.Printf("Server running at http://localhost%s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		fmt.Println("Server error:", err)
	}
}
