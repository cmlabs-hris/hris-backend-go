package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type leaveRequestRepositoryImpl struct {
	db *database.DB
}

// GetMyRequest implements leave.LeaveRequestRepository.
func (r *leaveRequestRepositoryImpl) GetMyRequest(ctx context.Context, userID string, companyID string) ([]leave.LeaveRequest, int64, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT lr.id, lr.employee_id, lr.leave_type_id, lr.start_date, lr.end_date, lr.duration_type, lr.total_days, lr.working_days, 
			   lr.reason, lr.attachment_url, lr.emergency_leave, lr.is_backdate, lr.status, lr.approved_by, lr.approved_at, 
			   lr.rejection_reason, lr.cancelled_by, lr.cancelled_at, lr.cancellation_reason, lr.submitted_at, lr.created_at, lr.updated_at
		FROM leave_requests lr
		INNER JOIN employees e ON lr.employee_id = e.id
		WHERE e.id = $1 AND e.company_id = $2
		ORDER BY lr.submitted_at DESC
	`

	rows, err := q.Query(ctx, query, userID, companyID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []leave.LeaveRequest
	for rows.Next() {
		var lr leave.LeaveRequest
		err := rows.Scan(
			&lr.ID,
			&lr.EmployeeID,
			&lr.LeaveTypeID,
			&lr.StartDate,
			&lr.EndDate,
			&lr.DurationType,
			&lr.TotalDays,
			&lr.WorkingDays,
			&lr.Reason,
			&lr.AttachmentURL,
			&lr.EmergencyLeave,
			&lr.IsBackdate,
			&lr.Status,
			&lr.ApprovedBy,
			&lr.ApprovedAt,
			&lr.RejectionReason,
			&lr.CancelledBy,
			&lr.CancelledAt,
			&lr.CancellationReason,
			&lr.SubmittedAt,
			&lr.CreatedAt,
			&lr.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, lr)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM leave_requests lr
		INNER JOIN employees e ON lr.employee_id = e.id
		WHERE e.id = $1 AND e.company_id = $2
	`
	var total int64
	err = q.QueryRow(ctx, countQuery, userID, companyID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// Delete implements leave.LeaveRequestRepository.
func (r *leaveRequestRepositoryImpl) Delete(ctx context.Context, id string) error {
	q := GetQuerier(ctx, r.db)
	query := `
		DELETE FROM leave_requests
		WHERE id = $1
	`
	commandTag, err := q.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("leave request with id %s not found", id)
	}
	return nil
}

func NewLeaveRequestRepository(db *database.DB) leave.LeaveRequestRepository {
	return &leaveRequestRepositoryImpl{db: db}
}

func (r *leaveRequestRepositoryImpl) Create(ctx context.Context, request leave.LeaveRequest) (leave.LeaveRequest, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO leave_requests (
			id, employee_id, leave_type_id,
			start_date, end_date, duration_type, total_days, working_days,
			reason, attachment_url, emergency_leave, is_backdate,
			status, submitted_at,
			created_at, updated_at
		) VALUES (
			uuidv7(), $1, $2,
			$3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, NOW(),
			NOW(), NOW()
		) RETURNING id, submitted_at, created_at, updated_at
	`

	err := q.QueryRow(ctx, query,
		request.EmployeeID, request.LeaveTypeID,
		request.StartDate, request.EndDate, request.DurationType, request.TotalDays, request.WorkingDays,
		request.Reason, request.AttachmentURL, request.EmergencyLeave, request.IsBackdate,
		request.Status,
	).Scan(&request.ID, &request.SubmittedAt, &request.CreatedAt, &request.UpdatedAt)

	if err != nil {
		return leave.LeaveRequest{}, nil
	}

	return request, nil
}

func (r *leaveRequestRepositoryImpl) GetByID(ctx context.Context, id string) (leave.LeaveRequest, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT lr.id, lr.employee_id, lr.leave_type_id,
			   lr.start_date, lr.end_date, lr.duration_type, lr.total_days, lr.working_days,
			   lr.reason, lr.attachment_url, lr.emergency_leave, lr.is_backdate,
			   lr.status,
			   lr.approved_by, lr.approved_at, lr.rejection_reason,
			   lr.cancelled_by, lr.cancelled_at, lr.cancellation_reason,
			   lr.submitted_at, lr.created_at, lr.updated_at,
			   lt.name as leave_type_name,
			   e.full_name as employee_name
		FROM leave_requests lr
		JOIN leave_types lt ON lr.leave_type_id = lt.id
		JOIN employees e ON lr.employee_id = e.id WHERE lr.id = $1
	`

	var req leave.LeaveRequest
	var leaveTypeName, employeeName string

	err := q.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.EmployeeID, &req.LeaveTypeID,
		&req.StartDate, &req.EndDate, &req.DurationType, &req.TotalDays, &req.WorkingDays,
		&req.Reason, &req.AttachmentURL, &req.EmergencyLeave, &req.IsBackdate,
		&req.Status,
		&req.ApprovedBy, &req.ApprovedAt, &req.RejectionReason,
		&req.CancelledBy, &req.CancelledAt, &req.CancellationReason,
		&req.SubmittedAt, &req.CreatedAt, &req.UpdatedAt,
		&leaveTypeName, &employeeName,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return leave.LeaveRequest{}, leave.ErrLeaveRequestNotFound
		}
		return leave.LeaveRequest{}, err
	}

	req.LeaveTypeName = &leaveTypeName
	req.EmployeeName = &employeeName

	return req, nil
}

func (r *leaveRequestRepositoryImpl) GetByEmployeeID(ctx context.Context, employeeID string, filter leave.LeaveRequestFilter) ([]leave.LeaveRequest, int64, error) {
	q := GetQuerier(ctx, r.db)

	// Build WHERE clause
	whereClause := "WHERE lr.employee_id = $1"
	args := []interface{}{employeeID}
	argIndex := 2

	if filter.LeaveTypeID != nil {
		whereClause += fmt.Sprintf(" AND lr.leave_type_id = $%d", argIndex)
		args = append(args, *filter.LeaveTypeID)
		argIndex++
	}

	if filter.Status != nil {
		whereClause += fmt.Sprintf(" AND lr.status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.StartDate != nil {
		whereClause += fmt.Sprintf(" AND lr.start_date >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		whereClause += fmt.Sprintf(" AND lr.end_date <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf(`
        SELECT COUNT(*) FROM leave_requests lr %s
    `, whereClause)

	var total int64
	err := q.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get data with pagination
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}

	offset := (filter.Page - 1) * filter.Limit

	query := fmt.Sprintf(`
		SELECT lr.id, lr.employee_id, lr.leave_type_id,
			   lr.start_date, lr.end_date, lr.duration_type, lr.total_days, lr.working_days,
			   lr.reason, lr.attachment_url, lr.emergency_leave, lr.is_backdate,
			   lr.status,
			   lr.approved_by, lr.approved_at, lr.rejection_reason,
			   lr.cancelled_by, lr.cancelled_at, lr.cancellation_reason,
			   lr.submitted_at, lr.created_at, lr.updated_at,
			   lt.name as leave_type_name,
			   e.full_name as employee_name
		FROM leave_requests lr
		JOIN leave_types lt ON lr.leave_type_id = lt.id
		JOIN employees e ON lr.employee_id = e.id
		%s
		ORDER BY lr.submitted_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.Limit, offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []leave.LeaveRequest

	for rows.Next() {
		var req leave.LeaveRequest
		var leaveTypeName, employeeName string

		err := rows.Scan(
			&req.ID, &req.EmployeeID, &req.LeaveTypeID,
			&req.StartDate, &req.EndDate, &req.DurationType, &req.TotalDays, &req.WorkingDays,
			&req.Reason, &req.AttachmentURL, &req.EmergencyLeave, &req.IsBackdate,
			&req.Status,
			&req.ApprovedBy, &req.ApprovedAt, &req.RejectionReason,
			&req.CancelledBy, &req.CancelledAt, &req.CancellationReason,
			&req.SubmittedAt, &req.CreatedAt, &req.UpdatedAt,
			&leaveTypeName, &employeeName,
		)

		if err != nil {
			return nil, 0, err
		}

		req.LeaveTypeName = &leaveTypeName
		req.EmployeeName = &employeeName

		requests = append(requests, req)
	}

	return requests, total, nil
}

func (r *leaveRequestRepositoryImpl) Update(ctx context.Context, request leave.UpdateLeaveRequestRequest) error {
	q := GetQuerier(ctx, r.db)

	updates := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	// Update fields based on UpdateLeaveRequestRequest
	if request.EmployeeID != nil {
		updates = append(updates, fmt.Sprintf("employee_id = $%d", argIdx))
		args = append(args, *request.EmployeeID)
		argIdx++
	}
	if request.LeaveTypeID != nil {
		updates = append(updates, fmt.Sprintf("leave_type_id = $%d", argIdx))
		args = append(args, *request.LeaveTypeID)
		argIdx++
	}
	if request.StartDate != nil {
		updates = append(updates, fmt.Sprintf("start_date = $%d", argIdx))
		args = append(args, *request.StartDate)
		argIdx++
	}
	if request.EndDate != nil {
		updates = append(updates, fmt.Sprintf("end_date = $%d", argIdx))
		args = append(args, *request.EndDate)
		argIdx++
	}
	if request.DurationType != nil {
		updates = append(updates, fmt.Sprintf("duration_type = $%d", argIdx))
		args = append(args, *request.DurationType)
		argIdx++
	}
	if request.TotalDays != nil {
		updates = append(updates, fmt.Sprintf("total_days = $%d", argIdx))
		args = append(args, *request.TotalDays)
		argIdx++
	}
	if request.WorkingDays != nil {
		updates = append(updates, fmt.Sprintf("working_days = $%d", argIdx))
		args = append(args, *request.WorkingDays)
		argIdx++
	}
	if request.Reason != nil {
		updates = append(updates, fmt.Sprintf("reason = $%d", argIdx))
		args = append(args, *request.Reason)
		argIdx++
	}
	if request.AttachmentURL != nil {
		updates = append(updates, fmt.Sprintf("attachment_url = $%d", argIdx))
		args = append(args, *request.AttachmentURL)
		argIdx++
	}
	if request.EmergencyLeave != nil {
		updates = append(updates, fmt.Sprintf("emergency_leave = $%d", argIdx))
		args = append(args, *request.EmergencyLeave)
		argIdx++
	}
	if request.IsBackdate != nil {
		updates = append(updates, fmt.Sprintf("is_backdate = $%d", argIdx))
		args = append(args, *request.IsBackdate)
		argIdx++
	}
	if request.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *request.Status)
		argIdx++
	}
	if request.ApprovedBy != nil {
		updates = append(updates, fmt.Sprintf("approved_by = $%d", argIdx))
		args = append(args, *request.ApprovedBy)
		argIdx++
	}
	if request.ApprovedAt != nil {
		updates = append(updates, fmt.Sprintf("approved_at = $%d", argIdx))
		args = append(args, *request.ApprovedAt)
		argIdx++
	}
	if request.RejectionReason != nil {
		updates = append(updates, fmt.Sprintf("rejection_reason = $%d", argIdx))
		args = append(args, *request.RejectionReason)
		argIdx++
	}
	if request.CancelledBy != nil {
		updates = append(updates, fmt.Sprintf("cancelled_by = $%d", argIdx))
		args = append(args, *request.CancelledBy)
		argIdx++
	}
	if request.CancelledAt != nil {
		updates = append(updates, fmt.Sprintf("cancelled_at = $%d", argIdx))
		args = append(args, *request.CancelledAt)
		argIdx++
	}
	if request.CancellationReason != nil {
		updates = append(updates, fmt.Sprintf("cancellation_reason = $%d", argIdx))
		args = append(args, *request.CancellationReason)
		argIdx++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updatable fields provided for leave request update")
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, request.ID)

	sql := "UPDATE leave_requests SET " + strings.Join(updates, ", ") + fmt.Sprintf(" WHERE id = $%d RETURNING id", argIdx)

	var updatedID string
	if err := q.QueryRow(ctx, sql, args...).Scan(&updatedID); err != nil {
		return fmt.Errorf("failed to update leave request with id %s: %w", request.ID, err)
	}
	return nil
}

func (r *leaveRequestRepositoryImpl) CheckOverlapping(
	ctx context.Context,
	employeeID string,
	startDate, endDate time.Time,
) (bool, error) {
	q := GetQuerier(ctx, r.db)

	query := `
        SELECT EXISTS (
            SELECT 1
            FROM leave_requests
            WHERE employee_id = $1
            AND status IN ('waiting_approval', 'approved')
            AND (
                (start_date <= $2 AND end_date >= $2) OR
                (start_date <= $3 AND end_date >= $3) OR
                (start_date >= $2 AND end_date <= $3)
            )
        )
    `

	var exists bool
	err := q.QueryRow(ctx, query, employeeID, startDate, endDate).Scan(&exists)

	return exists, err
}

// before

func (r *leaveRequestRepositoryImpl) GetByCompanyID(
	ctx context.Context,
	companyID string,
	filter leave.LeaveRequestFilter,
) ([]leave.LeaveRequest, int64, error) {
	q := GetQuerier(ctx, r.db)

	// Base query with JOINs
	baseQuery := `
        FROM leave_requests lr
        INNER JOIN employees e ON lr.employee_id = e.id
        INNER JOIN leave_types lt ON lr.leave_type_id = lt.id
        WHERE e.company_id = $1
    `

	args := []interface{}{companyID}
	argIdx := 2

	// Build WHERE clause dynamically
	whereClauses := []string{}

	// Filter by employee name (ILIKE for case-insensitive search)
	if filter.EmployeeName != nil && *filter.EmployeeName != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("e.full_name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.EmployeeName+"%")
		argIdx++
	}

	// Filter by employee ID
	if filter.EmployeeID != nil && *filter.EmployeeID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("lr.employee_id = $%d", argIdx))
		args = append(args, *filter.EmployeeID)
		argIdx++
	}

	// Filter by leave type
	if filter.LeaveTypeID != nil && *filter.LeaveTypeID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("lr.leave_type_id = $%d", argIdx))
		args = append(args, *filter.LeaveTypeID)
		argIdx++
	}

	// Filter by status
	if filter.Status != nil && *filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("lr.status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	// Filter by date range
	if filter.StartDate != nil && *filter.StartDate != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("lr.start_date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil && *filter.EndDate != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("lr.end_date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	// Append WHERE clauses
	if len(whereClauses) > 0 {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
	}

	// COUNT query for total records
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	err := q.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count leave requests: %w", err)
	}

	// Main SELECT query
	selectQuery := `
        SELECT 
            lr.id, lr.employee_id, lr.leave_type_id,
            lr.start_date, lr.end_date, lr.duration_type, lr.total_days, lr.working_days,
            lr.reason, lr.attachment_url, lr.emergency_leave, lr.is_backdate,
            lr.status,
            lr.approved_by, lr.approved_at, lr.rejection_reason,
            lr.cancelled_by, lr.cancelled_at, lr.cancellation_reason,
            lr.submitted_at, lr.created_at, lr.updated_at,
            lt.name as leave_type_name,
            e.full_name as employee_name,
    ` + baseQuery

	// ORDER BY clause
	orderBy := "lr.submitted_at DESC" // Default

	switch filter.SortBy {
	case "employee_name":
		orderBy = "e.full_name"
	case "start_date":
		orderBy = "lr.start_date"
	case "end_date":
		orderBy = "lr.end_date"
	case "status":
		orderBy = "lr.status"
	default:
		orderBy = "lr.submitted_at"
	}

	if strings.ToLower(filter.SortOrder) == "asc" {
		orderBy += " ASC"
	} else {
		orderBy += " DESC"
	}

	selectQuery += " ORDER BY " + orderBy

	// PAGINATION
	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit

	selectQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := q.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query leave requests: %w", err)
	}
	defer rows.Close()

	var requests []leave.LeaveRequest

	for rows.Next() {
		var req leave.LeaveRequest
		var leaveTypeName, employeeName string

		err := rows.Scan(
			&req.ID, &req.EmployeeID, &req.LeaveTypeID,
			&req.StartDate, &req.EndDate, &req.DurationType, &req.TotalDays, &req.WorkingDays,
			&req.Reason, &req.AttachmentURL, &req.EmergencyLeave, &req.IsBackdate,
			&req.Status,
			&req.ApprovedBy, &req.ApprovedAt, &req.RejectionReason,
			&req.CancelledBy, &req.CancelledAt, &req.CancellationReason,
			&req.SubmittedAt, &req.CreatedAt, &req.UpdatedAt,
			&leaveTypeName, &employeeName,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan leave request: %w", err)
		}

		req.LeaveTypeName = &leaveTypeName
		req.EmployeeName = &employeeName

		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return requests, total, nil
}

// GetMyRequests implements leave.LeaveRequestRepository with filtering for authenticated user
func (r *leaveRequestRepositoryImpl) GetMyRequests(ctx context.Context, employeeID string, companyID string, filter leave.MyLeaveRequestFilter) ([]leave.LeaveRequest, int64, error) {
	q := GetQuerier(ctx, r.db)

	// Build WHERE clause dynamically
	whereClauses := []string{"lr.employee_id = $1", "e.company_id = $2"}
	args := []interface{}{employeeID, companyID}
	paramCount := 2

	// Leave type filter
	if filter.LeaveTypeID != nil && *filter.LeaveTypeID != "" {
		paramCount++
		whereClauses = append(whereClauses, fmt.Sprintf("lr.leave_type_id = $%d", paramCount))
		args = append(args, *filter.LeaveTypeID)
	}

	// Status filter
	if filter.Status != nil && *filter.Status != "" {
		paramCount++
		whereClauses = append(whereClauses, fmt.Sprintf("lr.status = $%d", paramCount))
		args = append(args, *filter.Status)
	}

	// Date filters
	if filter.StartDate != nil && *filter.StartDate != "" {
		paramCount++
		whereClauses = append(whereClauses, fmt.Sprintf("lr.start_date >= $%d", paramCount))
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil && *filter.EndDate != "" {
		paramCount++
		whereClauses = append(whereClauses, fmt.Sprintf("lr.end_date <= $%d", paramCount))
		args = append(args, *filter.EndDate)
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Build ORDER BY clause
	var orderBy string
	switch filter.SortBy {
	case "start_date":
		orderBy = "lr.start_date"
	case "end_date":
		orderBy = "lr.end_date"
	case "status":
		orderBy = "lr.status"
	default:
		orderBy = "lr.submitted_at"
	}
	orderBy += " " + strings.ToUpper(filter.SortOrder)

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM leave_requests lr
		INNER JOIN employees e ON lr.employee_id = e.id
		WHERE %s
	`, whereClause)

	var total int64
	err := q.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count my leave requests: %w", err)
	}

	// Main query with JOINs
	offset := (filter.Page - 1) * filter.Limit
	query := fmt.Sprintf(`
		SELECT 
			lr.id, lr.employee_id, lr.leave_type_id, lr.start_date, lr.end_date,
			lr.duration_type, lr.total_days, lr.working_days, lr.reason, lr.attachment_url,
			lr.emergency_leave, lr.is_backdate, lr.status, lr.approved_by, lr.approved_at,
			lr.rejection_reason, lr.cancelled_by, lr.cancelled_at, lr.cancellation_reason,
			lr.submitted_at, lr.created_at, lr.updated_at,
			lt.name as leave_type_name,
			e.full_name as employee_name
		FROM leave_requests lr
		INNER JOIN employees e ON lr.employee_id = e.id
		INNER JOIN leave_types lt ON lr.leave_type_id = lt.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, paramCount+1, paramCount+2)

	args = append(args, filter.Limit, offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query my leave requests: %w", err)
	}
	defer rows.Close()

	var requests []leave.LeaveRequest
	for rows.Next() {
		var req leave.LeaveRequest
		var leaveTypeName, employeeName string

		err := rows.Scan(
			&req.ID, &req.EmployeeID, &req.LeaveTypeID, &req.StartDate, &req.EndDate,
			&req.DurationType, &req.TotalDays, &req.WorkingDays, &req.Reason, &req.AttachmentURL,
			&req.EmergencyLeave, &req.IsBackdate, &req.Status, &req.ApprovedBy, &req.ApprovedAt,
			&req.RejectionReason, &req.CancelledBy, &req.CancelledAt, &req.CancellationReason,
			&req.SubmittedAt, &req.CreatedAt, &req.UpdatedAt,
			&leaveTypeName, &employeeName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan leave request row: %w", err)
		}

		req.LeaveTypeName = &leaveTypeName
		req.EmployeeName = &employeeName
		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return requests, total, nil
}

func (r *leaveRequestRepositoryImpl) UpdateStatus(ctx context.Context, id, status string, approvedBy string) error {
	q := GetQuerier(ctx, r.db)
	var query string
	var args []interface{}
	if approvedBy != "" {
		query = `
			UPDATE leave_requests
			SET status = $1, approved_by = $2, approved_at = $3
			WHERE id = $4
			RETURNING id
		`
		args = []interface{}{status, approvedBy, time.Now(), id}
	} else {
		query = `
			UPDATE leave_requests
			SET status = $1
			WHERE id = $2
			RETURNING id
		`
		args = []interface{}{status, id}
	}
	var updatedID string
	err := q.QueryRow(ctx, query, args...).Scan(&updatedID)
	if err != nil {
		return fmt.Errorf("failed to update status for leave request with id %s: %w", id, err)
	}
	return nil
}
