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

type leaveQuotaRepositoryImpl struct {
	db *database.DB
}

// GetByCompanyID implements leave.LeaveQuotaRepository.
func (r *leaveQuotaRepositoryImpl) GetByCompanyID(ctx context.Context, companyID string) ([]leave.LeaveQuota, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT lq.id, lq.employee_id, lq.leave_type_id, lq.year,
			   lq.opening_balance, lq.earned_quota, lq.rollover_quota, lq.adjustment_quota,
			   lq.used_quota, lq.pending_quota, lq.available_quota, lq.rollover_expiry_date,
			   lq.created_at, lq.updated_at,
			   e.full_name AS employee_name,
			   lt.name AS leave_type_name
		FROM leave_quotas lq
		JOIN employees e ON lq.employee_id = e.id
		JOIN leave_types lt ON lq.leave_type_id = lt.id
		WHERE e.company_id = $1
		ORDER BY lq.year DESC, lt.name
	`

	rows, err := q.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotas := make([]leave.LeaveQuota, 0)
	for rows.Next() {
		var quota leave.LeaveQuota
		if err := rows.Scan(
			&quota.ID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
			&quota.OpeningBalance, &quota.EarnedQuota, &quota.RolloverQuota, &quota.AdjustmentQuota,
			&quota.UsedQuota, &quota.PendingQuota, &quota.AvailableQuota, &quota.RolloverExpiryDate,
			&quota.CreatedAt, &quota.UpdatedAt,
			&quota.EmployeeName, &quota.LeaveTypeName,
		); err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}

	return quotas, nil
}

// GetByCompanyIDAndYear implements leave.LeaveQuotaRepository.
func (r *leaveQuotaRepositoryImpl) GetByCompanyIDAndYear(ctx context.Context, companyID string, year int) ([]leave.LeaveQuota, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, employee_id, leave_type_id, year,
			   opening_balance, earned_quota, rollover_quota, adjustment_quota,
			   used_quota, pending_quota, available_quota, rollover_expiry_date,
			   created_at, updated_at
		FROM leave_quotas
		WHERE company_id = $1 AND year = $2
		ORDER BY leave_type_id
	`

	rows, err := q.Query(ctx, query, companyID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotas := make([]leave.LeaveQuota, 0)
	for rows.Next() {
		var quota leave.LeaveQuota
		if err := rows.Scan(
			&quota.ID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
			&quota.OpeningBalance, &quota.EarnedQuota, &quota.RolloverQuota, &quota.AdjustmentQuota,
			&quota.UsedQuota, &quota.PendingQuota, &quota.AvailableQuota, &quota.RolloverExpiryDate,
			&quota.CreatedAt, &quota.UpdatedAt,
		); err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}

	return quotas, nil
}

// GetByEmployee implements leave.LeaveQuotaRepository.
func (r *leaveQuotaRepositoryImpl) GetByEmployee(ctx context.Context, employeeID string) ([]leave.LeaveQuota, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, employee_id, leave_type_id, year,
			   opening_balance, earned_quota, rollover_quota, adjustment_quota,
			   used_quota, pending_quota, available_quota, rollover_expiry_date,
			   created_at, updated_at
		FROM leave_quotas
		WHERE employee_id = $1
		ORDER BY year DESC, leave_type_id
	`

	rows, err := q.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotas := make([]leave.LeaveQuota, 0)
	for rows.Next() {
		var quota leave.LeaveQuota
		if err := rows.Scan(
			&quota.ID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
			&quota.OpeningBalance, &quota.EarnedQuota, &quota.RolloverQuota, &quota.AdjustmentQuota,
			&quota.UsedQuota, &quota.PendingQuota, &quota.AvailableQuota, &quota.RolloverExpiryDate,
			&quota.CreatedAt, &quota.UpdatedAt,
		); err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}

	return quotas, nil
}

// GetByID implements leave.LeaveQuotaRepository.
func (r *leaveQuotaRepositoryImpl) GetByID(ctx context.Context, id string) (leave.LeaveQuota, error) {
	q := GetQuerier(ctx, r.db)
	query := `
		SELECT id, employee_id, leave_type_id, year,
			   opening_balance, earned_quota, rollover_quota, adjustment_quota,
			   used_quota, pending_quota, available_quota, rollover_expiry_date,
			   created_at, updated_at
		FROM leave_quotas
		WHERE id = $1
	`
	var quota leave.LeaveQuota
	err := q.QueryRow(ctx, query, id).Scan(
		&quota.ID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
		&quota.OpeningBalance, &quota.EarnedQuota, &quota.RolloverQuota, &quota.AdjustmentQuota,
		&quota.UsedQuota, &quota.PendingQuota, &quota.AvailableQuota, &quota.RolloverExpiryDate,
		&quota.CreatedAt, &quota.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return leave.LeaveQuota{}, leave.ErrQuotaNotFound
		}
		return leave.LeaveQuota{}, err
	}
	return quota, nil
}

// Delete implements leave.LeaveQuotaRepository.
func (r *leaveQuotaRepositoryImpl) Delete(ctx context.Context, id string) error {
	q := GetQuerier(ctx, r.db)
	query := `
		DELETE FROM leave_quotas
		WHERE id = $1
	`
	commandTag, err := q.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("leave quota with id %s not found", id)
	}
	return nil
}

func NewLeaveQuotaRepository(db *database.DB) leave.LeaveQuotaRepository {
	return &leaveQuotaRepositoryImpl{db: db}
}

// Create implements leave.LeaveQuotaRepository.
func (r *leaveQuotaRepositoryImpl) Create(ctx context.Context, quota leave.LeaveQuota) (leave.LeaveQuota, error) {
	q := GetQuerier(ctx, r.db)
	query := `
        INSERT INTO leave_quotas (
            id, employee_id, leave_type_id, year,
            opening_balance, earned_quota, rollover_quota, adjustment_quota,
            used_quota, pending_quota, rollover_expiry_date,
            created_at, updated_at
        ) VALUES (
            uuidv7(), $1, $2, $3,
            $4, $5, $6, $7,
            $8, $9, $10,
            NOW(), NOW()
        ) RETURNING id, available_quota, created_at, updated_at
    `

	err := q.QueryRow(ctx, query,
		quota.EmployeeID, quota.LeaveTypeID, quota.Year,
		quota.OpeningBalance, quota.EarnedQuota, quota.RolloverQuota, quota.AdjustmentQuota,
		quota.UsedQuota, quota.PendingQuota, quota.RolloverExpiryDate,
	).Scan(&quota.ID, &quota.AvailableQuota, &quota.CreatedAt, &quota.UpdatedAt)

	if err != nil {
		return leave.LeaveQuota{}, err
	}

	return quota, nil
}

func (r *leaveQuotaRepositoryImpl) GetByEmployeeTypeYear(ctx context.Context, employeeID, leaveTypeID string, year int) (leave.LeaveQuota, error) {
	q := GetQuerier(ctx, r.db)

	query := `
        SELECT id, employee_id, leave_type_id, year,
               opening_balance, earned_quota, rollover_quota, adjustment_quota,
               used_quota, pending_quota, available_quota, rollover_expiry_date,
               created_at, updated_at
        FROM leave_quotas
        WHERE employee_id = $1 AND leave_type_id = $2 AND year = $3
    `

	var quota leave.LeaveQuota

	err := q.QueryRow(ctx, query, employeeID, leaveTypeID, year).Scan(
		&quota.ID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
		&quota.OpeningBalance, &quota.EarnedQuota, &quota.RolloverQuota, &quota.AdjustmentQuota,
		&quota.UsedQuota, &quota.PendingQuota, &quota.AvailableQuota, &quota.RolloverExpiryDate,
		&quota.CreatedAt, &quota.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return leave.LeaveQuota{}, leave.ErrQuotaNotFound
		}
		return leave.LeaveQuota{}, err
	}

	return quota, nil

}

func (r *leaveQuotaRepositoryImpl) Update(ctx context.Context, req leave.UpdateLeaveQuotaRequest) error {
	q := GetQuerier(ctx, r.db)

	updates := make(map[string]interface{})

	if req.EmployeeID != nil {
		updates["employee_id"] = *req.EmployeeID
	}
	if req.LeaveTypeID != nil {
		updates["leave_type_id"] = *req.LeaveTypeID
	}
	if req.Year != nil {
		updates["year"] = *req.Year
	}
	if req.OpeningBalance != nil {
		updates["opening_balance"] = *req.OpeningBalance
	}
	if req.EarnedQuota != nil {
		updates["earned_quota"] = *req.EarnedQuota
	}
	if req.RolloverQuota != nil {
		updates["rollover_quota"] = *req.RolloverQuota
	}
	if req.AdjustmentQuota != nil {
		updates["adjustment_quota"] = *req.AdjustmentQuota
	}
	if req.UsedQuota != nil {
		updates["used_quota"] = *req.UsedQuota
	}
	if req.PendingQuota != nil {
		updates["pending_quota"] = *req.PendingQuota
	}
	if req.RolloverExpiryDate != nil {
		updates["rollover_expiry_date"] = *req.RolloverExpiryDate
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updatable fields provided for leave quota update")
	}
	updates["updated_at"] = time.Now()

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	sql := "UPDATE leave_quotas SET " +
		strings.Join(setClauses, ", ") +
		fmt.Sprintf(" WHERE id = $%d", i)
	args = append(args, req.ID)

	var updatedID string
	if err := q.QueryRow(ctx, sql+" RETURNING id", args...).Scan(&updatedID); err != nil {
		return fmt.Errorf("failed to update leave quota with id %s: %w", req.ID, err)
	}
	return nil
}

func (r *leaveQuotaRepositoryImpl) AddPendingQuota(ctx context.Context, quotaID string, amount float64) error {
	q := GetQuerier(ctx, r.db)

	query := `
    UPDATE leave_quotas
    SET pending_quota = pending_quota + $1,
        updated_at = NOW()
    WHERE id = $2
    AND (opening_balance + earned_quota + rollover_quota + adjustment_quota - used_quota - pending_quota - $1) >= 0
`

	result, err := q.Exec(ctx, query, amount, quotaID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return leave.ErrInsufficientQuota
	}

	return nil

}

func (r *leaveQuotaRepositoryImpl) RemovePendingQuota(ctx context.Context, quotaID string, amount float64) error {
	q := GetQuerier(ctx, r.db)

	query := `
    UPDATE leave_quotas
    SET pending_quota = pending_quota - $1,
        updated_at = NOW()
    WHERE id = $2
`

	_, err := q.Exec(ctx, query, amount, quotaID)
	return err

}

func (r *leaveQuotaRepositoryImpl) GetByEmployeeYear(ctx context.Context, employeeID string, year int) ([]leave.LeaveQuota, error) {
	q := GetQuerier(ctx, r.db)

	query := `
    SELECT lq.id, lq.employee_id, lq.leave_type_id, lq.year,
           lq.opening_balance, lq.earned_quota, lq.rollover_quota, lq.adjustment_quota,
           lq.used_quota, lq.pending_quota, lq.available_quota, lq.rollover_expiry_date,
           lq.created_at, lq.updated_at,
           lt.name as leave_type_name
    FROM leave_quotas lq
    JOIN leave_types lt ON lq.leave_type_id = lt.id
    WHERE lq.employee_id = $1 AND lq.year = $2
    ORDER BY lt.name
`

	rows, err := q.Query(ctx, query, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotas []leave.LeaveQuota

	for rows.Next() {
		var quota leave.LeaveQuota
		var leaveTypeName string

		err := rows.Scan(
			&quota.ID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
			&quota.OpeningBalance, &quota.EarnedQuota, &quota.RolloverQuota, &quota.AdjustmentQuota,
			&quota.UsedQuota, &quota.PendingQuota, &quota.AvailableQuota, &quota.RolloverExpiryDate,
			&quota.CreatedAt, &quota.UpdatedAt,
			&leaveTypeName,
		)

		if err != nil {
			return nil, err
		}

		quotas = append(quotas, quota)
	}

	return quotas, nil

}

func (r *leaveQuotaRepositoryImpl) MovePendingToUsed(ctx context.Context, quotaID string, amount float64) error {
	q := GetQuerier(ctx, r.db)

	query := `
    UPDATE leave_quotas
    SET pending_quota = pending_quota - $1,
        used_quota = used_quota + $1,
        updated_at = NOW()
    WHERE id = $2
`

	_, err := q.Exec(ctx, query, amount, quotaID)
	return err

}

// DecrementQuota implements leave.LeaveQuotaRepository.
func (r *leaveQuotaRepositoryImpl) DecrementQuota(ctx context.Context, quotaID string, days int) error {
	q := GetQuerier(ctx, r.db)
	query := `
		UPDATE leave_quotas 
		SET taken_quota = taken_quota + $1, updated_at = $3
		WHERE id = $2
		AND taken_quota + $1 <= total_quota
	`
	result, err := q.Exec(ctx, query, days, quotaID, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return leave.ErrInsufficientQuota
	}

	return nil
}
