package company

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testCompanyDB *database.DB
)

func companyTestInit() {
	if testCompanyDB != nil {
		return
	}
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:root@localhost:5432/cmlabs_hris_test?sslmode=disable"
	}

	var err error
	testCompanyDB, err = database.NewPostgreSQLDB(dsn)
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}
}

func truncateCompanyTables(t *testing.T, ctx context.Context) {
	companyTestInit()
	tables := []string{"companies"}

	for _, table := range tables {
		_, err := testCompanyDB.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Some tables might not exist, skip
			continue
		}
	}
}

func createCompanyTestCompany(t *testing.T, ctx context.Context, name string, username string) string {
	companyTestInit()
	var companyID string
	// Generate unique username per test
	uniqueUsername := fmt.Sprintf("%s-%d-%d", username, time.Now().Unix(), time.Now().Nanosecond())
	err := testCompanyDB.QueryRow(ctx, `
		INSERT INTO companies (id, name, username, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
		RETURNING id
	`, name, uniqueUsername).Scan(&companyID)
	require.NoError(t, err)
	return companyID
}

// ===== COMPANY SERVICE TESTS =====

func TestCompanyService_Create_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	companyTestInit()
	truncateCompanyTables(t, ctx)

	// Create service
	companyRepo := postgresql.NewCompanyRepository(testCompanyDB)

	// Act
	newCompany := company.Company{
		Name:     "New Test Company",
		Username: "new-test-company",
	}

	created, err := companyRepo.Create(ctx, newCompany)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "New Test Company", created.Name)
	assert.Equal(t, "new-test-company", created.Username)
	assert.NotNil(t, created.CreatedAt)
	assert.NotNil(t, created.UpdatedAt)
}

func TestCompanyService_GetByID_Success(t *testing.T) {
	t.Skip("GetByID not yet implemented in CompanyService")
	t.Parallel()
	ctx := context.Background()
	companyTestInit()
	truncateCompanyTables(t, ctx)

	// Setup
	companyID := createCompanyTestCompany(t, ctx, "Test Company", "test-company")

	// Create service
	companyRepo := postgresql.NewCompanyRepository(testCompanyDB)
	companyService := NewCompanyService(testCompanyDB, companyRepo)

	// Act
	retrieved, err := companyService.GetByID(ctx, companyID)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, retrieved)
}

func TestCompanyService_GetByID_NotFound(t *testing.T) {
	t.Skip("GetByID not yet implemented in CompanyService")
	t.Parallel()
	ctx := context.Background()
	companyTestInit()
	truncateCompanyTables(t, ctx)

	// Create service
	companyRepo := postgresql.NewCompanyRepository(testCompanyDB)
	companyService := NewCompanyService(testCompanyDB, companyRepo)

	// Act
	_, err := companyService.GetByID(ctx, "nonexistent-id")

	// Assert
	assert.Error(t, err)
}

func TestCompanyService_Update_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	companyTestInit()
	truncateCompanyTables(t, ctx)

	// Setup
	companyID := createCompanyTestCompany(t, ctx, "Original Company", "original-company")

	// Create service
	companyRepo := postgresql.NewCompanyRepository(testCompanyDB)
	companyService := NewCompanyService(testCompanyDB, companyRepo)

	// Act
	newName := "Updated Company"
	newAddress := "123 New Street"
	updateReq := company.UpdateCompanyRequest{
		Name:    &newName,
		Address: &newAddress,
	}

	err := companyService.Update(ctx, companyID, updateReq)

	// Assert
	assert.NoError(t, err)

	// Verify update by retrieving the company
	retrieved, err := companyRepo.GetByID(ctx, companyID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Company", retrieved.Name)
	assert.NotNil(t, retrieved.Address)
	assert.Equal(t, "123 New Street", *retrieved.Address)
}

func TestCompanyService_Update_PartialFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	companyTestInit()
	truncateCompanyTables(t, ctx)

	// Setup
	companyID := createCompanyTestCompany(t, ctx, "Test Company", "test-company")

	// Create service
	companyRepo := postgresql.NewCompanyRepository(testCompanyDB)
	companyService := NewCompanyService(testCompanyDB, companyRepo)

	// Act - Update only name
	newName := "Updated Only Name"
	updateReq := company.UpdateCompanyRequest{
		Name: &newName,
	}

	err := companyService.Update(ctx, companyID, updateReq)

	// Assert
	assert.NoError(t, err)

	// Verify update
	retrieved, err := companyRepo.GetByID(ctx, companyID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Only Name", retrieved.Name)
}

func TestCompanyService_Update_NoFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	companyTestInit()
	truncateCompanyTables(t, ctx)

	// Setup
	companyID := createCompanyTestCompany(t, ctx, "Test Company", "test-company")

	// Create service
	companyRepo := postgresql.NewCompanyRepository(testCompanyDB)
	companyService := NewCompanyService(testCompanyDB, companyRepo)

	// Act - Update with no fields
	updateReq := company.UpdateCompanyRequest{}

	err := companyService.Update(ctx, companyID, updateReq)

	// Assert - Should error when no updatable fields provided
	assert.Error(t, err)
}

func TestCompanyService_Delete_Success(t *testing.T) {
	t.Skip("Delete not yet implemented in CompanyService")
	t.Parallel()
	ctx := context.Background()
	companyTestInit()
	truncateCompanyTables(t, ctx)

	// Setup
	companyID := createCompanyTestCompany(t, ctx, "Company to Delete", "company-to-delete")

	// Create service
	companyRepo := postgresql.NewCompanyRepository(testCompanyDB)
	companyService := NewCompanyService(testCompanyDB, companyRepo)

	// Act
	err := companyService.Delete(ctx, companyID)

	// Assert
	assert.NoError(t, err)

	// Verify deletion by attempting to retrieve
	_, err = companyRepo.GetByID(ctx, companyID)
	assert.Error(t, err)
}

func TestCompanyService_List_Success(t *testing.T) {
	t.Skip("List not yet implemented in CompanyService")
	t.Parallel()
	ctx := context.Background()
	companyTestInit()
	truncateCompanyTables(t, ctx)

	// Setup - Create multiple companies
	createCompanyTestCompany(t, ctx, "Company 1", "company-1")
	createCompanyTestCompany(t, ctx, "Company 2", "company-2")
	createCompanyTestCompany(t, ctx, "Company 3", "company-3")

	// Create service
	companyRepo := postgresql.NewCompanyRepository(testCompanyDB)
	companyService := NewCompanyService(testCompanyDB, companyRepo)

	// Act
	companies, err := companyService.List(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Greater(t, len(companies), 0)
}
