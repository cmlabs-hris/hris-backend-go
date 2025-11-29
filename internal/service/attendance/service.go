package attendance

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/branch"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
)

type AttendanceServiceImpl struct {
	db *database.DB
	attendance.AttendanceRepository
	employee.EmployeeRepository
	schedule.WorkScheduleRepository
	schedule.WorkScheduleTimeRepository
	branch.BranchRepository
	fileService file.FileService
}

// timePtrToString safely converts a *time.Time to a string.
func timePtrToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	format := t.Format("2006-01-02 15:04:05")
	return &format
}

// ClockIn implements attendance.AttendanceService.
func (a *AttendanceServiceImpl) ClockIn(ctx context.Context, req attendance.ClockInRequest) (attendance.AttendanceResponse, error) {
	if err := req.Validate(); err != nil {
		return attendance.AttendanceResponse{}, err
	}
	nowUTC := time.Now().UTC()

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	employeeID, ok := claims["employee_id"].(string)
	if !ok || companyID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	timezoneStr, err := a.BranchRepository.GetTimezoneByEmployeeID(ctx, employeeID, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return attendance.AttendanceResponse{}, attendance.ErrNoScheduleFound
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get timezone by employee ID: %w", err)
	}

	loc, err := time.LoadLocation(timezoneStr)
	if err != nil {
		loc = time.UTC
	}

	nowLocal := nowUTC.In(loc)
	dateLocal := nowLocal.Format("2006-01-02")

	hasChekedIn, err := a.AttendanceRepository.HasCheckedInToday(ctx, employeeID, dateLocal, companyID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return attendance.AttendanceResponse{}, fmt.Errorf("failed to check if employee has checked in today: %w", err)
		}
	}

	if hasChekedIn {
		return attendance.AttendanceResponse{}, attendance.ErrAlreadyCheckedIn
	}

	activeSchedule, err := a.WorkScheduleRepository.GetActiveSchedule(ctx, employeeID, nowLocal, companyID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return attendance.AttendanceResponse{}, fmt.Errorf("failed to get active schedule: %w", err)
		}
	}

	if activeSchedule == nil {
		return attendance.AttendanceResponse{}, attendance.ErrNoScheduleFound
	}

	// if activeSchedule.LocationType != "WFA" {
	// 	isValidLocation := false

	// 	for _, office := range activeSchedule.Locations {
	// 		// Hitung jarak user ke kantor (dalam Meter)
	// 		distanceMeters := utils.CalculateHaversineDistance(
	// 			req.Latitude, req.Longitude,
	// 			office.Latitude, office.Longitude,
	// 		)

	// 		// Cek apakah masuk radius
	// 		if distanceMeters <= float64(office.RadiusMeters) {
	// 			isValidLocation = true
	// 			break // Keluar loop jika sudah ketemu satu yang valid
	// 		}
	// 	}

	// 	if !isValidLocation {
	// 		return attendance.AttendanceResponse{}, attendance.ErrOutsideAllowedRadius
	// 	}
	// }

	scheduledInTime := time.Date(
		nowLocal.Year(), nowLocal.Month(), nowLocal.Day(),
		activeSchedule.ClockIn.Hour(), activeSchedule.ClockIn.Minute(), 0, 0,
		loc,
	)

	// Batas Toleransi (Grace Period)
	graceLimitTime := scheduledInTime.Add(time.Duration(activeSchedule.GracePeriodMinutes) * time.Minute)

	status := "PRESENT"
	lateMinutes := 0

	// Logika Terlambat
	if nowLocal.After(graceLimitTime) {
		status = "LATE"
		// Hitung selisih dari Jadwal Asli (bukan dari grace period)
		diff := nowLocal.Sub(scheduledInTime).Minutes()
		if diff > 0 {
			lateMinutes = int(math.Floor(diff))
		}
	}
	status = "waiting_approval"

	// Validasi Early Check-In (Opsional: Cegah absen jam 2 pagi untuk shift jam 8)
	// Misal: Max checkin 3 jam sebelum jadwal
	earliestAllowed := scheduledInTime.Add(-1 * time.Hour)
	if nowLocal.Before(earliestAllowed) {
		return attendance.AttendanceResponse{}, attendance.ErrTooEarlyToCheckIn
	}

	ProofPhotoURL, err := a.fileService.UploadAttendanceProof(ctx, employeeID, nowLocal.Truncate(24*time.Hour), req.File, req.FileHeader.Filename, "CLOCK_IN")
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to upload attendance proof: %w", err)
	}
	req.ProofPhotoURL = &ProofPhotoURL

	data := attendance.Attendance{
		EmployeeID: req.EmployeeID,
		CompanyID:  companyID,

		// PENTING: Date adalah representasi "Hari Kerja", bukan timestamp
		Date: nowLocal.Truncate(24 * time.Hour), // time.Time (trunc to day)

		// Referensi ke Rule Jadwal
		WorkScheduleTimeID: &activeSchedule.TimeID,
		ActualLocationType: &activeSchedule.LocationType, // WFO/WFA/Hybrid

		// Waktu Absolut (Simpan UTC!)
		ClockIn: &nowUTC,

		// Bukti Lokasi
		ClockInLatitude:  &req.Latitude,
		ClockInLongitude: &req.Longitude,
		ClockInProofURL:  req.ProofPhotoURL,

		// Hasil Kalkulasi
		Status:            status,
		LateMinutes:       &lateMinutes,
		EarlyLeaveMinutes: nil, // Diisi saat checkout
		OvertimeMinutes:   nil, // Diisi saat checkout
	}

	attendanceResult, err := a.AttendanceRepository.Create(ctx, data)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to create attendance record: %w", err)
	}

	return attendance.AttendanceResponse{
		ID:                attendanceResult.ID,
		EmployeeID:        attendanceResult.EmployeeID,
		EmployeeName:      *attendanceResult.EmployeeName,
		Date:              attendanceResult.Date.Format("2006-01-02"),
		ClockInTime:       timePtrToString(attendanceResult.ClockIn),
		ClockOutTime:      timePtrToString(attendanceResult.ClockOut),
		ClockInLatitude:   attendanceResult.ClockInLatitude,
		ClockInLongitude:  attendanceResult.ClockInLongitude,
		ClockOutLatitude:  attendanceResult.ClockOutLatitude,
		ClockOutLongitude: attendanceResult.ClockOutLongitude,
		ClockInProofURL:   attendanceResult.ClockInProofURL,
		ClockOutProofURL:  attendanceResult.ClockOutProofURL,
		WorkingHours:      nil,
		Status:            attendanceResult.Status,
		IsLate:            nil,
		IsEarlyLeave:      nil,
		LateMinutes:       attendanceResult.LateMinutes,
		EarlyLeaveMinutes: attendanceResult.EarlyLeaveMinutes,
	}, nil
}

// ClockOut implements attendance.AttendanceService.
func (a *AttendanceServiceImpl) ClockOut(ctx context.Context, req attendance.ClockOutRequest) (attendance.AttendanceResponse, error) {
	if err := req.Validate(); err != nil {
		return attendance.AttendanceResponse{}, err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	employeeID, ok := claims["employee_id"].(string)
	if !ok || companyID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	attendanceData, err := a.AttendanceRepository.GetOpenSession(ctx, employeeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return attendance.AttendanceResponse{}, attendance.ErrNotCheckedIn
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get open session: %w", err)
	}

	if attendanceData.WorkScheduleTimeID == nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("attendance has no associated work schedule time")
	}

	scheduleTime, err := a.WorkScheduleTimeRepository.GetByID(ctx, *attendanceData.WorkScheduleTimeID, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return attendance.AttendanceResponse{}, schedule.ErrWorkScheduleTimeNotFound
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get work schedule time: %w", err)
	}

	timezoneStr, err := a.BranchRepository.GetTimezoneByEmployeeID(ctx, employeeID, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return attendance.AttendanceResponse{}, branch.ErrBranchNotFound
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get branch location: %w", err)
	}

	loc, err := time.LoadLocation(timezoneStr)
	if err != nil {
		loc = time.UTC
	}
	nowUTC := time.Now().UTC()
	nowLocal := nowUTC.In(loc)

	scheduledOut := time.Date(
		attendanceData.Date.Year(), attendanceData.Date.Month(), attendanceData.Date.Day(),
		scheduleTime.ClockOutTime.Hour(), scheduleTime.ClockOutTime.Minute(), 0, 0,
		loc,
	)

	// Jika Shift Malam, tambah 1 hari
	if scheduleTime.IsNextDayCheckout {
		scheduledOut = scheduledOut.Add(24 * time.Hour)
	}

	// 5. Kalkulasi Selisih (Dalam Menit)
	var earlyLeaveMins int
	var overtimeMins int

	// Cek Pulang Cepat
	if nowUTC.Before(scheduledOut) {
		diff := scheduledOut.Sub(nowUTC).Minutes()
		earlyLeaveMins = int(diff)
	}

	// Hitung Total Jam Kerja
	workDuration := nowUTC.Sub(*attendanceData.ClockIn)
	workHoursMins := int(workDuration.Minutes())

	ProofPhotoURL, err := a.fileService.UploadAttendanceProof(ctx, employeeID, nowLocal.Truncate(24*time.Hour), req.File, req.FileHeader.Filename, "CLOCK_OUT")
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to upload attendance proof: %w", err)
	}
	req.ProofPhotoURL = &ProofPhotoURL

	attendanceData.ClockOut = &nowUTC
	attendanceData.ClockOutLatitude = &req.Latitude
	attendanceData.ClockOutLongitude = &req.Longitude
	attendanceData.EarlyLeaveMinutes = &earlyLeaveMins
	attendanceData.OvertimeMinutes = &overtimeMins
	attendanceData.WorkHoursInMinutes = &workHoursMins
	attendanceData.ClockOutProofURL = req.ProofPhotoURL

	if err := a.AttendanceRepository.Update(ctx, attendanceData); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return attendance.AttendanceResponse{}, fmt.Errorf("attendance not found: %w", attendance.ErrAttendanceNotFound)
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to update attendance record: %w", err)
	}

	return attendance.AttendanceResponse{
		ID:                attendanceData.ID,
		EmployeeID:        attendanceData.EmployeeID,
		EmployeeName:      *attendanceData.EmployeeName,
		Date:              attendanceData.Date.Format("2006-01-02"),
		ClockInTime:       timePtrToString(attendanceData.ClockIn),
		ClockOutTime:      timePtrToString(attendanceData.ClockOut),
		ClockInLatitude:   attendanceData.ClockInLatitude,
		ClockInLongitude:  attendanceData.ClockInLongitude,
		ClockOutLatitude:  attendanceData.ClockOutLatitude,
		ClockOutLongitude: attendanceData.ClockOutLongitude,
		ClockInProofURL:   attendanceData.ClockInProofURL,
		ClockOutProofURL:  attendanceData.ClockOutProofURL,
		WorkingHours:      func(v float64) *float64 { return &v }(workDuration.Minutes()),
		Status:            attendanceData.Status,
		IsLate:            nil,
		IsEarlyLeave:      nil,
		LateMinutes:       attendanceData.LateMinutes,
		EarlyLeaveMinutes: attendanceData.EarlyLeaveMinutes,
	}, nil
}

// GetMyAttendance implements attendance.AttendanceService.
func (a *AttendanceServiceImpl) GetMyAttendance(ctx context.Context, filter attendance.MyAttendanceFilter) (attendance.ListAttendanceResponse, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return attendance.ListAttendanceResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return attendance.ListAttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	employeeID, ok := claims["employee_id"].(string)
	if !ok || employeeID == "" {
		return attendance.ListAttendanceResponse{}, fmt.Errorf("employee_id claim is missing or invalid")
	}

	attendances, total, err := a.AttendanceRepository.GetMyAttendance(ctx, employeeID, filter, companyID)
	if err != nil {
		return attendance.ListAttendanceResponse{}, fmt.Errorf("failed to get my attendance: %w", err)
	}

	// Map to response
	responses := make([]attendance.AttendanceResponse, 0, len(attendances))
	for _, att := range attendances {
		responses = append(responses, mapAttendanceToResponse(att))
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))
	showing := fmt.Sprintf("%d-%d of %d", (filter.Page-1)*filter.Limit+1, min((filter.Page)*filter.Limit, int(total)), total)
	if total == 0 {
		showing = "0 of 0"
	}

	return attendance.ListAttendanceResponse{
		TotalCount:  total,
		Page:        filter.Page,
		Limit:       filter.Limit,
		TotalPages:  totalPages,
		Showing:     showing,
		Attendances: responses,
	}, nil
}

// ListAttendance implements attendance.AttendanceService.
func (a *AttendanceServiceImpl) ListAttendance(ctx context.Context, filter attendance.AttendanceFilter) (attendance.ListAttendanceResponse, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return attendance.ListAttendanceResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return attendance.ListAttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	attendances, total, err := a.AttendanceRepository.List(ctx, filter, companyID)
	if err != nil {
		return attendance.ListAttendanceResponse{}, fmt.Errorf("failed to list attendances: %w", err)
	}

	// Map to response
	responses := make([]attendance.AttendanceResponse, 0, len(attendances))
	for _, att := range attendances {
		responses = append(responses, mapAttendanceToResponse(att))
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))
	showing := fmt.Sprintf("%d-%d of %d", (filter.Page-1)*filter.Limit+1, min((filter.Page)*filter.Limit, int(total)), total)
	if total == 0 {
		showing = "0 of 0"
	}

	return attendance.ListAttendanceResponse{
		TotalCount:  total,
		Page:        filter.Page,
		Limit:       filter.Limit,
		TotalPages:  totalPages,
		Showing:     showing,
		Attendances: responses,
	}, nil
}

// mapAttendanceToResponse converts an Attendance entity to AttendanceResponse
func mapAttendanceToResponse(att attendance.Attendance) attendance.AttendanceResponse {
	var employeeName string
	if att.EmployeeName != nil {
		employeeName = *att.EmployeeName
	}

	var workingHours *float64
	if att.WorkHoursInMinutes != nil {
		hours := float64(*att.WorkHoursInMinutes) / 60.0
		workingHours = &hours
	}

	var isLate *bool
	if att.LateMinutes != nil && *att.LateMinutes > 0 {
		v := true
		isLate = &v
	}

	var isEarlyLeave *bool
	if att.EarlyLeaveMinutes != nil && *att.EarlyLeaveMinutes > 0 {
		v := true
		isEarlyLeave = &v
	}

	return attendance.AttendanceResponse{
		ID:                att.ID,
		EmployeeID:        att.EmployeeID,
		EmployeeName:      employeeName,
		EmployeePosition:  att.EmployeePosition,
		Date:              att.Date.Format("2006-01-02"),
		ClockInTime:       timePtrToString(att.ClockIn),
		ClockOutTime:      timePtrToString(att.ClockOut),
		ClockInLatitude:   att.ClockInLatitude,
		ClockInLongitude:  att.ClockInLongitude,
		ClockOutLatitude:  att.ClockOutLatitude,
		ClockOutLongitude: att.ClockOutLongitude,
		ClockInProofURL:   att.ClockInProofURL,
		ClockOutProofURL:  att.ClockOutProofURL,
		WorkingHours:      workingHours,
		Status:            att.Status,
		IsLate:            isLate,
		IsEarlyLeave:      isEarlyLeave,
		LateMinutes:       att.LateMinutes,
		EarlyLeaveMinutes: att.EarlyLeaveMinutes,
		CreatedAt:         att.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:         att.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// UpdateAttendance implements attendance.AttendanceService.
// This allows managers/owners to fix attendance data like wrong clock times, etc.
func (a *AttendanceServiceImpl) UpdateAttendance(ctx context.Context, req attendance.UpdateAttendanceRequest) (attendance.AttendanceResponse, error) {
	if err := req.Validate(); err != nil {
		return attendance.AttendanceResponse{}, err
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	// Get existing attendance
	att, err := a.AttendanceRepository.GetByID(ctx, req.ID, companyID)
	if err != nil {
		if errors.Is(err, attendance.ErrAttendanceNotFound) {
			return attendance.AttendanceResponse{}, attendance.ErrAttendanceNotFound
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get attendance: %w", err)
	}

	// Update fields based on request
	if req.Date != nil && *req.Date != "" {
		parsedDate, _ := time.Parse("2006-01-02", *req.Date)
		att.Date = parsedDate
	}

	if req.ClockInTime != nil && *req.ClockInTime != "" {
		// Try parsing as full datetime first, then as time only
		clockIn, err := time.Parse("2006-01-02 15:04:05", *req.ClockInTime)
		if err != nil {
			// Try parsing as time only and combine with attendance date
			clockInTime, err := time.Parse("15:04:05", *req.ClockInTime)
			if err == nil {
				clockIn = time.Date(att.Date.Year(), att.Date.Month(), att.Date.Day(),
					clockInTime.Hour(), clockInTime.Minute(), clockInTime.Second(), 0, time.UTC)
			}
		}
		if !clockIn.IsZero() {
			att.ClockIn = &clockIn
		}
	}

	if req.ClockOutTime != nil && *req.ClockOutTime != "" {
		// Try parsing as full datetime first, then as time only
		clockOut, err := time.Parse("2006-01-02 15:04:05", *req.ClockOutTime)
		if err != nil {
			// Try parsing as time only and combine with attendance date
			clockOutTime, err := time.Parse("15:04:05", *req.ClockOutTime)
			if err == nil {
				clockOut = time.Date(att.Date.Year(), att.Date.Month(), att.Date.Day(),
					clockOutTime.Hour(), clockOutTime.Minute(), clockOutTime.Second(), 0, time.UTC)
			}
		}
		if !clockOut.IsZero() {
			att.ClockOut = &clockOut
		}
	}

	if req.ClockInLatitude != nil {
		att.ClockInLatitude = req.ClockInLatitude
	}
	if req.ClockInLongitude != nil {
		att.ClockInLongitude = req.ClockInLongitude
	}
	if req.ClockOutLatitude != nil {
		att.ClockOutLatitude = req.ClockOutLatitude
	}
	if req.ClockOutLongitude != nil {
		att.ClockOutLongitude = req.ClockOutLongitude
	}

	if req.Status != nil {
		att.Status = *req.Status
	}

	if req.LateMinutes != nil {
		att.LateMinutes = req.LateMinutes
	}
	if req.EarlyLeaveMinutes != nil {
		att.EarlyLeaveMinutes = req.EarlyLeaveMinutes
	}
	if req.OvertimeMinutes != nil {
		att.OvertimeMinutes = req.OvertimeMinutes
	}

	// Recalculate work hours if both clock in and out are present
	if att.ClockIn != nil && att.ClockOut != nil {
		workDuration := att.ClockOut.Sub(*att.ClockIn)
		workHoursMins := int(workDuration.Minutes())
		att.WorkHoursInMinutes = &workHoursMins
	}

	// Update in repository
	if err := a.AttendanceRepository.Update(ctx, att); err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to update attendance: %w", err)
	}

	// Fetch updated record
	updatedAtt, err := a.AttendanceRepository.GetByID(ctx, req.ID, companyID)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get updated attendance: %w", err)
	}

	return mapAttendanceToResponse(updatedAtt), nil
}

// GetAttendance implements attendance.AttendanceService.
func (a *AttendanceServiceImpl) GetAttendance(ctx context.Context, id string) (attendance.AttendanceResponse, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	att, err := a.AttendanceRepository.GetByID(ctx, id, companyID)
	if err != nil {
		if errors.Is(err, attendance.ErrAttendanceNotFound) {
			return attendance.AttendanceResponse{}, attendance.ErrAttendanceNotFound
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get attendance: %w", err)
	}

	return mapAttendanceToResponse(att), nil
}

// ApproveAttendance implements attendance.AttendanceService.
func (a *AttendanceServiceImpl) ApproveAttendance(ctx context.Context, req attendance.ApproveAttendanceRequest) (attendance.AttendanceResponse, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("user_id claim is missing or invalid")
	}

	// Get existing attendance
	att, err := a.AttendanceRepository.GetByID(ctx, req.ID, companyID)
	if err != nil {
		if errors.Is(err, attendance.ErrAttendanceNotFound) {
			return attendance.AttendanceResponse{}, attendance.ErrAttendanceNotFound
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get attendance: %w", err)
	}

	// Update status and approver info
	now := time.Now()
	att.Status = "approved"
	att.ApprovedBy = &userID
	att.ApprovedAt = &now
	att.RejectionReason = nil // Clear any rejection reason

	// Update in repository
	if err := a.AttendanceRepository.Update(ctx, att); err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to approve attendance: %w", err)
	}

	// Fetch updated record
	updatedAtt, err := a.AttendanceRepository.GetByID(ctx, req.ID, companyID)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get updated attendance: %w", err)
	}

	return mapAttendanceToResponse(updatedAtt), nil
}

// RejectAttendance implements attendance.AttendanceService.
func (a *AttendanceServiceImpl) RejectAttendance(ctx context.Context, req attendance.RejectAttendanceRequest) (attendance.AttendanceResponse, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("company_id claim is missing or invalid")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return attendance.AttendanceResponse{}, fmt.Errorf("user_id claim is missing or invalid")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return attendance.AttendanceResponse{}, err
	}

	// Get existing attendance
	att, err := a.AttendanceRepository.GetByID(ctx, req.ID, companyID)
	if err != nil {
		if errors.Is(err, attendance.ErrAttendanceNotFound) {
			return attendance.AttendanceResponse{}, attendance.ErrAttendanceNotFound
		}
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get attendance: %w", err)
	}

	// Update status and approver info
	now := time.Now()
	att.Status = "rejected"
	att.ApprovedBy = &userID
	att.ApprovedAt = &now
	att.RejectionReason = &req.Reason

	// Update in repository
	if err := a.AttendanceRepository.Update(ctx, att); err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to reject attendance: %w", err)
	}

	// Fetch updated record
	updatedAtt, err := a.AttendanceRepository.GetByID(ctx, req.ID, companyID)
	if err != nil {
		return attendance.AttendanceResponse{}, fmt.Errorf("failed to get updated attendance: %w", err)
	}

	return mapAttendanceToResponse(updatedAtt), nil
}

// DeleteAttendance implements attendance.AttendanceService.
func (a *AttendanceServiceImpl) DeleteAttendance(ctx context.Context, id string) error {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return fmt.Errorf("company_id claim is missing or invalid")
	}

	if err := a.AttendanceRepository.Delete(ctx, id, companyID); err != nil {
		if errors.Is(err, attendance.ErrAttendanceNotFound) {
			return attendance.ErrAttendanceNotFound
		}
		return fmt.Errorf("failed to delete attendance: %w", err)
	}

	return nil
}

func NewAttendanceService(
	db *database.DB,
	attendanceRepo attendance.AttendanceRepository,
	employeeRepo employee.EmployeeRepository,
	workScheduleRepo schedule.WorkScheduleRepository,
	workScheduleTimeRepo schedule.WorkScheduleTimeRepository,
	branchRepo branch.BranchRepository,
	fileService file.FileService,
) attendance.AttendanceService {
	return &AttendanceServiceImpl{
		db:                         db,
		AttendanceRepository:       attendanceRepo,
		EmployeeRepository:         employeeRepo,
		WorkScheduleRepository:     workScheduleRepo,
		WorkScheduleTimeRepository: workScheduleTimeRepo,
		BranchRepository:           branchRepo,
		fileService:                fileService,
	}
}
