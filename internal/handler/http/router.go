package http

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/middleware"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/jwtauth/v5"
)

func NewRouter(JWTService jwt.Service, authHandler AuthHandler, companyhandler CompanyHandler, leaveHandler LeaveHandler, masterHandler MasterHandler, scheduleHandler ScheduleHandler, attendanceHandler AttendanceHandler, employeeHandler EmployeeHandler, invitationHandler InvitationHandler, payrollHandler PayrollHandler, dashboardHandler DashboardHandler, employeeDashboardHandler EmployeeDashboardHandler, storageBasePath string) *chi.Mux {
	r := chi.NewRouter()
	logFormat := httplog.SchemaECS.Concise(false)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	})).With(
		slog.String("app", "hris-cmlabs"),
		slog.String("version", "v1.0.0"),
		slog.String("env", "development"),
	)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		MaxAge:           300,
	}))

	// r.Use(chiMiddleware.RealIP)

	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:           slog.LevelDebug,
		Schema:          httplog.SchemaECS,
		LogRequestBody:  func(req *http.Request) bool { return true },
		LogResponseBody: func(req *http.Request) bool { return true },
	}))

	r.Use(chiMiddleware.AllowContentType("application/json", "multipart/form-data"))
	r.Use(chiMiddleware.CleanPath)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Heartbeat("/"))

	fileServer := http.FileServer(http.Dir(storageBasePath))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	r.Route("/api/v1", func(r chi.Router) {

		// Public invitation route (no auth required) - uses /view/ prefix to avoid conflict with /my
		r.Get("/invitations/view/{token}", invitationHandler.GetInvitationByToken)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/logout", authHandler.Logout)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Route("/oauth/callback", func(r chi.Router) {
				r.Get("/google", authHandler.OAuthCallbackGoogle)
			})

			r.Route("/login", func(r chi.Router) {
				r.Post("/", authHandler.Login)
				r.Post("/employee-code", authHandler.LoginWithEmployeeCode)
				r.Route("/oauth", func(r chi.Router) {
					r.Get("/google", authHandler.LoginWithGoogle)
				})
			})

		})

		// Requires authentication
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(JWTService.JWTAuth()))
			r.Use(middleware.AuthRequired(JWTService.JWTAuth()))
			r.Use(middleware.RequireCompany)

			r.Route("/company", func(r chi.Router) {

				// Pending only
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequirePending)
					r.Post("/", companyhandler.Create)
				})

				r.Route("/my", func(r chi.Router) {
					r.Get("/", companyhandler.GetByID)

					// Owner only
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireOwner)
						r.Put("/", companyhandler.Update)
						r.Delete("/", companyhandler.Delete)
						r.Post("/logo", companyhandler.UploadCompanyLogo)
					})
				})
			})

			r.Route("/leave", func(r chi.Router) {
				r.Route("/types", func(r chi.Router) {
					r.Get("/", leaveHandler.ListTypes)
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireOwner)
						r.Post("/", leaveHandler.CreateType)
						r.Put("/{id}", leaveHandler.UpdateType)
						r.Delete("/{id}", leaveHandler.DeleteType)
					})

				})

				r.Route("/quota", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireOwner)
						r.Get("/", leaveHandler.ListQuota)
						// r.Post("/", leaveHandler.SetQuota)
						r.Post("/adjust", leaveHandler.AdjustQuota)
					})
					r.Get("/my", leaveHandler.GetMyQuota)
					r.Get("/{id}", leaveHandler.GetQuota)
				})

				r.Route("/requests", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireManager)
						r.Get("/", leaveHandler.ListRequests)
						r.Post("/{id}/approve", leaveHandler.ApproveRequest)
						r.Post("/{id}/reject", leaveHandler.RejectRequest)
					})
					r.Post("/", leaveHandler.CreateRequest)
					r.Get("/{id}", leaveHandler.GetRequest)
					r.Get("/my", leaveHandler.GetMyRequests)
				})
			})

			// Master Data Routes
			r.Route("/master", func(r chi.Router) {
				// Branch routes
				r.Route("/branches", func(r chi.Router) {
					r.Get("/", masterHandler.ListBranches)
					r.Get("/{id}", masterHandler.GetBranch)

					// Owner/Manager only
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireManager)
						r.Post("/", masterHandler.CreateBranch)
						r.Put("/{id}", masterHandler.UpdateBranch)
						r.Delete("/{id}", masterHandler.DeleteBranch)
					})
				})

				// Grade routes
				r.Route("/grades", func(r chi.Router) {
					r.Get("/", masterHandler.ListGrades)
					r.Get("/{id}", masterHandler.GetGrade)

					// Owner/Manager only
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireManager)
						r.Post("/", masterHandler.CreateGrade)
						r.Put("/{id}", masterHandler.UpdateGrade)
						r.Delete("/{id}", masterHandler.DeleteGrade)
					})
				})

				// Position routes
				r.Route("/positions", func(r chi.Router) {
					r.Get("/", masterHandler.ListPositions)
					r.Get("/{id}", masterHandler.GetPosition)

					// Owner/Manager only
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireManager)
						r.Post("/", masterHandler.CreatePosition)
						r.Put("/{id}", masterHandler.UpdatePosition)
						r.Delete("/{id}", masterHandler.DeletePosition)
					})
				})
			})

			r.Route("/schedule", func(r chi.Router) {
				// Work Schedule
				r.Get("/", scheduleHandler.ListWorkSchedules)   // All
				r.Get("/{id}", scheduleHandler.GetWorkSchedule) // All

				r.Post("/{scheduleID}/employee/{employeeID}", scheduleHandler.AssignSchedule)
				r.Put("/{assignID}/employee/{employeeID}", scheduleHandler.UpdateEmployeeScheduleAssignment)
				r.Delete("/{assignID}/employee/{employeeID}", scheduleHandler.DeleteEmployeeScheduleAssignment)
				// Employee Schedule Timeline
				r.Get("/employee/{id}", scheduleHandler.GetEmployeeScheduleTimeline) // All - Get timeline for specific employee

				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireOwner)
					r.Post("/", scheduleHandler.CreateWorkSchedule)       // Owner
					r.Put("/{id}", scheduleHandler.UpdateWorkSchedule)    // Owner
					r.Delete("/{id}", scheduleHandler.DeleteWorkSchedule) // Owner
				})

				// Work Schedule Times
				r.Route("/times", func(r chi.Router) {
					r.Get("/{id}", scheduleHandler.GetWorkScheduleTime) // All
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireOwner)
						r.Post("/", scheduleHandler.CreateWorkScheduleTime)       // Owner
						r.Put("/{id}", scheduleHandler.UpdateWorkScheduleTime)    // Owner
						r.Delete("/{id}", scheduleHandler.DeleteWorkScheduleTime) // Owner
					})
				})

				// Work Schedule Locations
				r.Route("/locations", func(r chi.Router) {
					r.Get("/{id}", scheduleHandler.GetWorkScheduleLocation) // All
					r.Group(func(r chi.Router) {
						r.Post("/", scheduleHandler.CreateWorkScheduleLocation)       // Owner
						r.Put("/{id}", scheduleHandler.UpdateWorkScheduleLocation)    // Owner
						r.Delete("/{id}", scheduleHandler.DeleteWorkScheduleLocation) // Owner
					})
				})

			})

			// Employee Schedule Assignments
			r.Route("/employee-schedules", func(r chi.Router) {
				r.Get("/", scheduleHandler.ListEmployeeScheduleAssignments)    // All (filtered by employee_id)
				r.Get("/active", scheduleHandler.GetActiveScheduleForEmployee) // All
				r.Get("/{id}", scheduleHandler.GetEmployeeScheduleAssignment)  // All
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireManager)
					r.Post("/", scheduleHandler.CreateEmployeeScheduleAssignment)       // Owner
					r.Put("/{id}", scheduleHandler.UpdateEmployeeScheduleAssignment)    // Owner
					r.Delete("/{id}", scheduleHandler.DeleteEmployeeScheduleAssignment) // Owner
				})
			})

			// Attendance Routes
			r.Route("/attendance", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireManager)
					r.Get("/", attendanceHandler.List)                 // All with filters
					r.Get("/{id}", attendanceHandler.Get)              // Get single attendance
					r.Put("/{id}", attendanceHandler.Update)           // Update attendance (fix records)
					r.Delete("/{id}", attendanceHandler.Delete)        // Delete attendance
					r.Post("/{id}/approve", attendanceHandler.Approve) // Approve attendance
					r.Post("/{id}/reject", attendanceHandler.Reject)   // Reject attendance
				})
				r.Get("/my", attendanceHandler.GetMyAttendance)  // Get my attendance records
				r.Post("/clock-in", attendanceHandler.ClockIn)   // Clock in
				r.Post("/clock-out", attendanceHandler.ClockOut) // Clock out
			})

			r.Route("/employees", func(r chi.Router) {
				r.Get("/{id}", employeeHandler.GetEmployee) // Get single employee

				// Manager+ routes
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireManager)
					r.Get("/", employeeHandler.ListEmployees)                           // List employees with filters
					r.Get("/search", employeeHandler.SearchEmployees)                   // Autocomplete search
					r.Post("/", employeeHandler.CreateEmployee)                         // Create employee (multipart)
					r.Put("/{id}", employeeHandler.UpdateEmployee)                      // Update employee
					r.Delete("/{id}", employeeHandler.DeleteEmployee)                   // Soft delete employee
					r.Post("/{id}/inactivate", employeeHandler.InactivateEmployee)      // Inactivate employee
					r.Post("/{id}/invitation/resend", employeeHandler.ResendInvitation) // Resend invitation
					r.Post("/{id}/invitation/revoke", employeeHandler.RevokeInvitation) // Revoke invitation
				})

				r.Post("/{id}/avatar", employeeHandler.UploadAvatar) // Upload avatar
			})

			// Invitation Routes
			r.Route("/invitations", func(r chi.Router) {
				r.Get("/my", invitationHandler.ListMyInvitations)             // List pending invitations for current user
				r.Post("/{token}/accept", invitationHandler.AcceptInvitation) // Accept invitation
			})

			// Payroll Routes
			r.Route("/payroll", func(r chi.Router) {
				r.Use(middleware.RequireManager)

				// Settings
				r.Get("/settings", payrollHandler.GetSettings)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireOwner)
					r.Put("/settings", payrollHandler.UpdateSettings)
				})

				// Components
				r.Get("/components", payrollHandler.ListComponents)
				r.Get("/components/{id}", payrollHandler.GetComponent)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireOwner)
					r.Post("/components", payrollHandler.CreateComponent)
					r.Put("/components/{id}", payrollHandler.UpdateComponent)
					r.Delete("/components/{id}", payrollHandler.DeleteComponent)
				})

				// Employee Components
				r.Post("/employees/{employeeId}/components", payrollHandler.AssignComponent)
				r.Get("/employees/{employeeId}/components", payrollHandler.GetEmployeeComponents)
				r.Put("/employee-components/{id}", payrollHandler.UpdateEmployeeComponent)
				r.Delete("/employee-components/{id}", payrollHandler.RemoveEmployeeComponent)

				// Payroll Records
				r.Post("/generate", payrollHandler.GeneratePayroll)
				r.Get("/records", payrollHandler.ListPayrollRecords)
				r.Get("/records/{id}", payrollHandler.GetPayrollRecord)
				r.Put("/records/{id}", payrollHandler.UpdatePayrollRecord)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireOwner)
					r.Delete("/records/{id}", payrollHandler.DeletePayrollRecord)
					r.Post("/finalize", payrollHandler.FinalizePayroll)
				})

				// Summary
				r.Get("/summary", payrollHandler.GetPayrollSummary)
			})

			// Dashboard Routes (Manager+)
			r.Route("/dashboard", func(r chi.Router) {
				r.Route("/admin", func(r chi.Router) {
					r.Use(middleware.RequireManager)
					r.Get("/", dashboardHandler.GetDashboard)
					r.Get("/employee-current-number", dashboardHandler.GetEmployeeCurrentNumber)
					r.Get("/employee-status-stats", dashboardHandler.GetEmployeeStatusStats)
					r.Get("/monthly-attendance", dashboardHandler.GetMonthlyAttendance)
					r.Get("/daily-attendance-stats", dashboardHandler.GetDailyAttendanceStats)
				})
				r.Route("/employee", func(r chi.Router) {
					r.Get("/", employeeDashboardHandler.GetDashboard)
					r.Get("/work-stats", employeeDashboardHandler.GetWorkStats)
					r.Get("/attendance-summary", employeeDashboardHandler.GetAttendanceSummary)
					r.Get("/leave-summary", employeeDashboardHandler.GetLeaveSummary)
					r.Get("/work-hours-chart", employeeDashboardHandler.GetWorkHoursChart)
				})
			})

		})
	})
	return r
}
