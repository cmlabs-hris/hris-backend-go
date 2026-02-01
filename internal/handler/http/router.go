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

func NewRouter(JWTService jwt.Service, authHandler AuthHandler, companyhandler CompanyHandler, leaveHandler LeaveHandler, masterHandler MasterHandler, scheduleHandler ScheduleHandler, attendanceHandler AttendanceHandler, employeeHandler EmployeeHandler, invitationHandler InvitationHandler, payrollHandler PayrollHandler, dashboardHandler DashboardHandler, employeeDashboardHandler EmployeeDashboardHandler, notificationHandler NotificationHandler, reportHandler ReportHandler, subscriptionHandler SubscriptionHandler, subscriptionMiddleware *middleware.SubscriptionMiddleware, storageBasePath string) *chi.Mux {
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

		// Public subscription routes
		r.Get("/plans", subscriptionHandler.GetPlans)

		// Xendit webhook (public, signature verified)
		r.Post("/webhook/xendit", subscriptionHandler.HandleWebhook)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/logout", authHandler.Logout)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
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

				r.Group(func(r chi.Router) {
					r.Use(subscriptionMiddleware.RequireActiveSubscription)
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
			})

			r.Route("/leave", func(r chi.Router) {
				// Leave Types
				r.Route("/types", func(r chi.Router) {
					// Read operations - available to all subscriptions
					r.Get("/", leaveHandler.ListTypes)

					// Write operations - require leave feature
					r.Group(func(r chi.Router) {
						r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureLeave))
						r.Use(middleware.RequireOwner)
						r.Post("/", leaveHandler.CreateType)
						r.Put("/{id}", leaveHandler.UpdateType)
						r.Delete("/{id}", leaveHandler.DeleteType)
					})
				})

				// Leave Quota
				r.Route("/quota", func(r chi.Router) {
					// Read operations - available to all subscriptions
					r.Get("/my", leaveHandler.GetMyQuota)
					r.Get("/{id}", leaveHandler.GetQuota)

					// Manager read + write operations - require leave feature
					r.Group(func(r chi.Router) {
						r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureLeave))
						r.Use(middleware.RequireManager)
						r.Get("/", leaveHandler.ListQuota)
						r.Post("/adjust", leaveHandler.AdjustQuota)
					})
				})

				// Leave Requests
				r.Route("/requests", func(r chi.Router) {
					// Read operations - available to all subscriptions
					r.Get("/{id}", leaveHandler.GetRequest)
					r.Get("/my", leaveHandler.GetMyRequests)

					// Write operations - require leave feature
					r.Group(func(r chi.Router) {
						r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureLeave))
						r.Post("/", leaveHandler.CreateRequest)

						// Manager operations
						r.Group(func(r chi.Router) {
							r.Use(middleware.RequireManager)
							r.Get("/", leaveHandler.ListRequests)
							r.Post("/{id}/approve", leaveHandler.ApproveRequest)
							r.Post("/{id}/reject", leaveHandler.RejectRequest)
						})
					})
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
				// Read operations - available to all subscriptions (schedules are core system data)
				r.Get("/", scheduleHandler.ListWorkSchedules)
				r.Get("/{id}", scheduleHandler.GetWorkSchedule)
				r.Get("/employee/{id}", scheduleHandler.GetEmployeeScheduleTimeline)

				// Write operations - require schedule feature
				r.Group(func(r chi.Router) {
					r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureSchedule))

					r.Post("/{scheduleID}/employee/{employeeID}", scheduleHandler.AssignSchedule)
					r.Put("/{assignID}/employee/{employeeID}", scheduleHandler.UpdateEmployeeScheduleAssignment)
					r.Delete("/{assignID}/employee/{employeeID}", scheduleHandler.DeleteEmployeeScheduleAssignment)

					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireOwner)
						r.Post("/", scheduleHandler.CreateWorkSchedule)
						r.Put("/{id}", scheduleHandler.UpdateWorkSchedule)
						r.Delete("/{id}", scheduleHandler.DeleteWorkSchedule)
					})
				})

				// Work Schedule Times
				r.Route("/times", func(r chi.Router) {
					r.Get("/{id}", scheduleHandler.GetWorkScheduleTime) // Read - all subscriptions
					r.Group(func(r chi.Router) {
						r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureSchedule))
						r.Use(middleware.RequireOwner)
						r.Post("/", scheduleHandler.CreateWorkScheduleTime)
						r.Put("/{id}", scheduleHandler.UpdateWorkScheduleTime)
						r.Delete("/{id}", scheduleHandler.DeleteWorkScheduleTime)
					})
				})

				// Work Schedule Locations
				r.Route("/locations", func(r chi.Router) {
					r.Get("/{id}", scheduleHandler.GetWorkScheduleLocation) // Read - all subscriptions
					r.Group(func(r chi.Router) {
						r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureSchedule))
						r.Post("/", scheduleHandler.CreateWorkScheduleLocation)
						r.Put("/{id}", scheduleHandler.UpdateWorkScheduleLocation)
						r.Delete("/{id}", scheduleHandler.DeleteWorkScheduleLocation)
					})
				})

			})

			// Employee Schedule Assignments
			r.Route("/employee-schedules", func(r chi.Router) {
				// Read operations - available to all subscriptions
				r.Get("/", scheduleHandler.ListEmployeeScheduleAssignments)
				r.Get("/active", scheduleHandler.GetActiveScheduleForEmployee)
				r.Get("/{id}", scheduleHandler.GetEmployeeScheduleAssignment)

				// Write operations - require schedule feature
				r.Group(func(r chi.Router) {
					r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureSchedule))
					r.Use(middleware.RequireManager)
					r.Post("/", scheduleHandler.CreateEmployeeScheduleAssignment)
					r.Put("/{id}", scheduleHandler.UpdateEmployeeScheduleAssignment)
					r.Delete("/{id}", scheduleHandler.DeleteEmployeeScheduleAssignment)
				})
			})

			// Attendance Routes
			r.Route("/attendance", func(r chi.Router) {
				// Read operations - available to all subscriptions
				r.Get("/my", attendanceHandler.GetMyAttendance) // Get my attendance records

				// Write operations - require attendance feature
				r.Group(func(r chi.Router) {
					r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureAttendance))
					r.Post("/clock-in", attendanceHandler.ClockIn)   // Clock in
					r.Post("/clock-out", attendanceHandler.ClockOut) // Clock out

					// Manager operations
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireManager)
						r.Get("/", attendanceHandler.List)                 // All with filters
						r.Get("/{id}", attendanceHandler.Get)              // Get single attendance
						r.Put("/{id}", attendanceHandler.Update)           // Update attendance (fix records)
						r.Delete("/{id}", attendanceHandler.Delete)        // Delete attendance
						r.Post("/{id}/approve", attendanceHandler.Approve) // Approve attendance
						r.Post("/{id}/reject", attendanceHandler.Reject)   // Reject attendance
					})
				})
			})

			r.Route("/employees", func(r chi.Router) {
				r.Get("/{id}", employeeHandler.GetEmployee) // Get single employee

				// Manager+ routes (requires invitation feature for creating employees)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireManager)
					r.Get("/", employeeHandler.ListEmployees)         // List employees with filters
					r.Get("/search", employeeHandler.SearchEmployees) // Autocomplete search

					// Create employee requires invitation feature + employee slot check
					r.Group(func(r chi.Router) {
						r.Use(subscriptionMiddleware.RequireFeature(middleware.FeatureInvitation))
						r.Use(subscriptionMiddleware.RequireCanAddEmployee)
						r.Post("/", employeeHandler.CreateEmployee) // Create employee (multipart)
					})

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

				// Read operations - available to all subscriptions
				r.Get("/settings", payrollHandler.GetSettings)
				r.Get("/components", payrollHandler.ListComponents)
				r.Get("/components/{id}", payrollHandler.GetComponent)
				r.Get("/employees/{employeeId}/components", payrollHandler.GetEmployeeComponents)
				r.Get("/records", payrollHandler.ListPayrollRecords)
				r.Get("/records/{id}", payrollHandler.GetPayrollRecord)
				r.Get("/summary", payrollHandler.GetPayrollSummary)

				// Write operations - require payroll feature
				r.Group(func(r chi.Router) {
					r.Use(subscriptionMiddleware.RequireFeature(middleware.FeaturePayroll))

					// Settings (Owner only)
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireOwner)
						r.Put("/settings", payrollHandler.UpdateSettings)
					})

					// Components (Owner only)
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireOwner)
						r.Post("/components", payrollHandler.CreateComponent)
						r.Put("/components/{id}", payrollHandler.UpdateComponent)
						r.Delete("/components/{id}", payrollHandler.DeleteComponent)
					})

					// Employee Components
					r.Post("/employees/{employeeId}/components", payrollHandler.AssignComponent)
					r.Put("/employee-components/{id}", payrollHandler.UpdateEmployeeComponent)
					r.Delete("/employee-components/{id}", payrollHandler.RemoveEmployeeComponent)

					// Payroll Records
					r.Post("/generate", payrollHandler.GeneratePayroll)
					r.Put("/records/{id}", payrollHandler.UpdatePayrollRecord)
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireOwner)
						r.Delete("/records/{id}", payrollHandler.DeletePayrollRecord)
						r.Post("/finalize", payrollHandler.FinalizePayroll)
					})
				})
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

			// Notification Routes
			r.Route("/notifications", func(r chi.Router) {
				// Get SSE token (requires JWT auth)
				r.Get("/token", notificationHandler.GetSSEToken)

				// SSE Stream (uses SSE token from query param)
				r.Get("/stream", notificationHandler.Stream)

				// CRUD operations
				r.Get("/", notificationHandler.List)
				r.Get("/unread-count", notificationHandler.UnreadCount)
				r.Post("/mark-read", notificationHandler.MarkAsRead)
				r.Post("/mark-all-read", notificationHandler.MarkAllAsRead)
				r.Delete("/{id}", notificationHandler.Delete)

				// Preferences
				r.Get("/preferences", notificationHandler.GetPreferences)
				r.Put("/preferences", notificationHandler.UpdatePreference)
			})

			// Report Routes (Manager+) - Read-only, available to all subscriptions
			r.Route("/reports", func(r chi.Router) {
				r.Use(middleware.RequireManager)

				r.Get("/attendance", reportHandler.GetMonthlyAttendanceReport)
				r.Get("/payroll", reportHandler.GetPayrollSummaryReport)
				r.Get("/leave-balance", reportHandler.GetLeaveBalanceReport)
				r.Get("/new-hires", reportHandler.GetNewHireReport)
			})

			// Subscription Routes
			r.Route("/subscription", func(r chi.Router) {
				// Authenticated routes - view subscription and invoices
				r.Get("/my", subscriptionHandler.GetMySubscription)
				r.Get("/invoices", subscriptionHandler.GetInvoices)
				r.Get("/invoices/{id}", subscriptionHandler.GetInvoiceByID)

				// Owner-only routes - manage subscription
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireOwner)
					r.Post("/checkout", subscriptionHandler.Checkout)
					r.Post("/upgrade", subscriptionHandler.UpgradePlan)
					r.Post("/downgrade", subscriptionHandler.DowngradePlan)
					r.Post("/cancel", subscriptionHandler.CancelSubscription)
					r.Post("/seats", subscriptionHandler.ChangeSeats)
					r.Delete("/invoices/{id}", subscriptionHandler.CancelPendingInvoice)
				})
			})

		})
	})
	return r
}
