package postgresql_test

import (
	"context"
	"fmt"
	"os"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

// TestDatabaseSetup untuk menginisialisasi test database
type TestDatabaseSetup struct {
	DB *database.DB
}

// NewTestDatabase membuat koneksi ke test database
func NewTestDatabase() (*TestDatabaseSetup, error) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:root@localhost:5432/cmlabs_hris_test?sslmode=disable"
	}

	db, err := database.NewPostgreSQLDB(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	return &TestDatabaseSetup{DB: db}, nil
}

// TruncateAllTables menghapus semua data dari tabel
func (t *TestDatabaseSetup) TruncateAllTables(ctx context.Context) error {
	tx, err := t.DB.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tables := []string{
		"users",
		"companies",
		"employees",
		"organization_units",
		"positions",
		"grades",
		"branches",
		"work_schedules",
		"work_schedule_times",
		"work_schedule_locations",
		"employee_schedule_assignments",
		"attendances",
		"leave_types",
		"leave_quotas",
		"leave_requests",
		"document_types",
		"document_templates",
		"employee_documents",
		"employee_job_history",
		"audit_trails",
	}

	for _, table := range tables {
		_, err := tx.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return tx.Commit(ctx)
}

// Close menutup koneksi database
func (t *TestDatabaseSetup) Close() {
	t.DB.Close()
}
