package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/config"
	appHTTP "github.com/cmlabs-hris/hris-backend-go/internal/handler/http"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/middleware"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/cron"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/email"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/oauth"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/sse"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/storage"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/xendit"
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
	notificationService "github.com/cmlabs-hris/hris-backend-go/internal/service/notification"
	payrollService "github.com/cmlabs-hris/hris-backend-go/internal/service/payroll"
	reportService "github.com/cmlabs-hris/hris-backend-go/internal/service/report"
	scheduleService "github.com/cmlabs-hris/hris-backend-go/internal/service/schedule"
	subscriptionService "github.com/cmlabs-hris/hris-backend-go/internal/service/subscription"
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
	passwordResetRepo := postgresql.NewPasswordResetRepository(db)
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
	notificationRepo := postgresql.NewNotificationRepository(db)
	reportRepo := postgresql.NewReportRepository(db)

	// Subscription repositories
	featureRepo := postgresql.NewFeatureRepository(db)
	planRepo := postgresql.NewPlanRepository(db)
	subscriptionRepo := postgresql.NewSubscriptionRepository(db)
	invoiceRepo := postgresql.NewInvoiceRepository(db)
	employeeCounter := postgresql.NewEmployeeCounter(db)

	// Initialize SSE Hub for real-time notifications
	sseHub := sse.NewHub()

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

	// Initialize Xendit client
	xenditClient := xendit.NewClient(cfg.Xendit)
	webhookVerifier := xendit.NewWebhookVerifier(cfg.Xendit.WebhookToken)

	// Initialize subscription service
	subscriptionSvc := subscriptionService.NewSubscriptionService(
		featureRepo,
		planRepo,
		subscriptionRepo,
		invoiceRepo,
		employeeCounter,
		xenditClient,
		db,
		cfg,
	)

	// Initialize subscription middleware
	subscriptionMiddleware := middleware.NewSubscriptionMiddleware(subscriptionSvc)

	authService := serviceAuth.NewAuthService(db, userRepo, companyRepo, JWTService, JWTRepository, passwordResetRepo, employeeRepo, emailService, cfg.App.FrontendURL, subscriptionSvc)
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
		notificationRepo,
		subscriptionSvc,
	)
	masterService := master.NewMasterService(branchRepo, gradeRepo, positionRepo)
	notificationSvc := notificationService.NewNotificationService(notificationRepo, sseHub, notificationService.Config{
		BatchSize:     100,
		FlushInterval: 5 * time.Second,
		WorkerCount:   2,
		QueueSize:     1000,
	})
	leaveService := leave.NewLeaveService(db, leaveTypeRepo, leaveQuotaRepo, leaveRequestRepo, employeeRepo, attendanceRepo, quotaService, requestService, fileService, notificationSvc)
	scheduleService := scheduleService.NewScheduleService(
		db,
		workScheduleRepo,
		workScheduleTimeRepo,
		workScheduleLocationRepo,
		employeeScheduleAssignmentRepo,
		employeeRepo,
		notificationSvc,
	)
	attendanceService := attendanceService.NewAttendanceService(
		db,
		attendanceRepo,
		employeeRepo,
		workScheduleRepo,
		workScheduleTimeRepo,
		branchRepo,
		fileService,
		notificationSvc,
	)
	invitationService := invitationService.NewInvitationService(
		db,
		invitationRepo,
		employeeRepo,
		userRepo,
		emailService,
		fileService,
		cfg.Invitation,
		notificationSvc,
	)
	employeeService := employeeService.NewEmployeeService(
		db,
		employeeRepo,
		companyRepo,
		fileService,
		invitationService,
		quotaService,
		subscriptionSvc,
	)
	payrollSvc := payrollService.NewPayrollService(db, payrollRepo, employeeRepo, notificationSvc)
	dashboardSvc := dashboardService.NewDashboardService(dashboardRepo)
	empDashboardSvc := employeeDashboardService.NewEmployeeDashboardService(empDashboardRepo)
	reportSvc := reportService.NewReportService(reportRepo)

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
	notificationHandler := appHTTP.NewNotificationHandler(notificationSvc, JWTService)
	reportHandler := appHTTP.NewReportHandler(reportSvc)
	subscriptionHandler := appHTTP.NewSubscriptionHandler(subscriptionSvc, webhookVerifier)

	// Initialize cron scheduler
	cronScheduler := cron.NewScheduler()
	subscriptionJobs := cron.NewSubscriptionJobs(subscriptionSvc)
	subscriptionJobs.RegisterJobs(cronScheduler)
	attendanceJobs := cron.NewAttendanceJobs(
		attendanceRepo,
		employeeRepo,
		workScheduleRepo,
		workScheduleTimeRepo,
		branchRepo,
		notificationSvc,
		db,
	)
	attendanceJobs.RegisterJobs(cronScheduler)
	go cronScheduler.Start()
	defer cronScheduler.Stop()

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
		notificationHandler,
		reportHandler,
		subscriptionHandler,
		subscriptionMiddleware,
		cfg.Storage.BasePath,
	)

	port := fmt.Sprintf(":%d", cfg.App.Port)
	fmt.Printf("Server running at http://localhost%s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		fmt.Println("Server error:", err)
	}
}
