package postgresql_test

import (
	"context"
	"os"
	"testing"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var testDB *database.DB

func init() {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		// Fallback untuk local testing
		dsn = "postgres://postgres:postgres@localhost:5432/hris_test?sslmode=disable"
	}

	var err error
	testDB, err = database.NewPostgreSQLDB(dsn)
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}
}

// Setup function untuk membersihkan dan setup data test
func setupTestData(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	// Truncate tables
	_, err = tx.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)

	_, err = tx.Exec(ctx, "TRUNCATE TABLE companies CASCADE")
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)
}

// Cleanup function untuk reset data setelah test
func cleanupTestData(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)

	_, err = tx.Exec(ctx, "TRUNCATE TABLE companies CASCADE")
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)
}

// Helper untuk membuat company untuk testing
func createTestCompany(t *testing.T, ctx context.Context) string {
	var companyID string
	err := testDB.QueryRow(ctx, `
		INSERT INTO companies (id, name, username, created_at, updated_at)
		VALUES (gen_random_uuid(), 'Test Company', 'test-company', NOW(), NOW())
		RETURNING id
	`).Scan(&companyID)
	require.NoError(t, err)
	return companyID
}

// Helper untuk membuat user untuk testing
func createTestUser(t *testing.T, ctx context.Context, companyID string) user.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	hashedStr := string(hashedPassword)

	var newUser user.User
	err := testDB.QueryRow(ctx, `
		INSERT INTO users (id, company_id, email, password_hash, is_admin, email_verified, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'test@example.com', $2, false, true, NOW(), NOW())
		RETURNING id, company_id, email, password_hash, is_admin, oauth_provider, oauth_provider_id,
				  email_verified, email_verification_token, email_verification_sent_at,
				  created_at, updated_at
	`, companyID, hashedStr).Scan(
		&newUser.ID, &newUser.CompanyID, &newUser.Email, &newUser.PasswordHash, &newUser.IsAdmin,
		&newUser.OAuthProvider, &newUser.OAuthProviderID, &newUser.EmailVerified,
		&newUser.EmailVerificationToken, &newUser.EmailVerificationSentAt,
		&newUser.CreatedAt, &newUser.UpdatedAt,
	)
	require.NoError(t, err)
	return newUser
}

// ===== USER REPOSITORY TESTS =====

func TestUserRepository_Create_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyID := createTestCompany(t, ctx)
	userRepo := postgresql.NewUserRepository(testDB)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("securepass"), bcrypt.DefaultCost)
	hashedStr := string(hashedPassword)

	newUser := user.User{
		CompanyID:     &companyID,
		Email:         "newuser@example.com",
		PasswordHash:  &hashedStr,
		IsAdmin:       false,
		EmailVerified: true,
	}

	created, err := userRepo.Create(ctx, newUser)

	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, newUser.Email, created.Email)
	assert.Equal(t, newUser.IsAdmin, created.IsAdmin)
	assert.NotNil(t, created.CreatedAt)
	assert.NotNil(t, created.UpdatedAt)
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyID := createTestCompany(t, ctx)
	userRepo := postgresql.NewUserRepository(testDB)

	// Create test user
	testUser := createTestUser(t, ctx, companyID)

	// Get user by email
	retrieved, err := userRepo.GetByEmail(ctx, "test@example.com")

	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, retrieved.ID)
	assert.Equal(t, testUser.Email, retrieved.Email)
	assert.Equal(t, testUser.IsAdmin, retrieved.IsAdmin)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	userRepo := postgresql.NewUserRepository(testDB)

	_, err := userRepo.GetByEmail(ctx, "notfound@example.com")

	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyID := createTestCompany(t, ctx)
	userRepo := postgresql.NewUserRepository(testDB)

	testUser := createTestUser(t, ctx, companyID)

	retrieved, err := userRepo.GetByID(ctx, testUser.ID)

	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, retrieved.ID)
	assert.Equal(t, testUser.Email, retrieved.Email)
}

func TestUserRepository_LinkGoogleAccount_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyID := createTestCompany(t, ctx)
	userRepo := postgresql.NewUserRepository(testDB)

	testUser := createTestUser(t, ctx, companyID)

	// Link Google account
	linked, err := userRepo.LinkGoogleAccount(ctx, "google-id-123", testUser.Email)

	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, linked.ID)
	assert.NotNil(t, linked.OAuthProvider)
	assert.Equal(t, "google", *linked.OAuthProvider)
	assert.NotNil(t, linked.OAuthProviderID)
	assert.Equal(t, "google-id-123", *linked.OAuthProviderID)
}

func TestUserRepository_LinkPasswordAccount_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyID := createTestCompany(t, ctx)
	userRepo := postgresql.NewUserRepository(testDB)

	// Create user without password
	var userID string
	err := testDB.QueryRow(ctx, `
		INSERT INTO users (id, company_id, email, is_admin, email_verified, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'nopass@example.com', false, true, NOW(), NOW())
		RETURNING id
	`, companyID).Scan(&userID)
	require.NoError(t, err)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("newpassword123"), bcrypt.DefaultCost)

	linked, err := userRepo.LinkPasswordAccount(ctx, userID, string(hashedPassword))

	assert.NoError(t, err)
	assert.Equal(t, userID, linked.ID)
	assert.NotNil(t, linked.PasswordHash)
}

// func TestUserRepository_Update_Success(t *testing.T) {
// 	defer cleanupTestData(t)
// 	setupTestData(t)

// 	ctx := context.Background()
// 	companyID := createTestCompany(t, ctx)
// 	userRepo := postgresql.NewUserRepository(testDB)

// 	testUser := createTestUser(t, ctx, companyID)

// 	// Update user
// 	updateReq := user.UpdateUserRequest{
// 		Email:   strPtr("updated@example.com"),
// 		IsAdmin: boolPtr(true),
// 	}

// 	updated, err := userRepo.Update(ctx, testUser.ID, updateReq)

// 	assert.NoError(t, err)
// 	assert.Equal(t, testUser.ID, updated.ID)
// 	assert.Equal(t, "updated@example.com", updated.Email)
// 	assert.Equal(t, true, updated.IsAdmin)
// }

// func TestUserRepository_Delete_Success(t *testing.T) {
// 	defer cleanupTestData(t)
// 	setupTestData(t)

// 	ctx := context.Background()
// 	companyID := createTestCompany(t, ctx)
// 	userRepo := postgresql.NewUserRepository(testDB)

// 	testUser := createTestUser(t, ctx, companyID)

// 	// Delete user
// 	err := userRepo.Delete(ctx, testUser.ID)

// 	assert.NoError(t, err)

// 	// Verify deletion
// 	_, err = userRepo.GetByID(ctx, testUser.ID)
// 	assert.Error(t, err)
// }

// ===== COMPANY REPOSITORY TESTS =====

// func TestCompanyRepository_Create_Success(t *testing.T) {
// 	defer cleanupTestData(t)
// 	setupTestData(t)

// 	ctx := context.Background()
// 	companyRepo := postgresql.NewCompanyRepository(testDB)

// 	newCompany := user.Company{
// 		Name:     "New Company",
// 		Username: "new-company-" + time.Now().Format("20060102150405"),
// 		Address:  strPtr("123 Main St"),
// 	}

// 	created, err := companyRepo.Create(ctx, newCompany)

// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, created.ID)
// 	assert.Equal(t, newCompany.Name, created.Name)
// 	assert.Equal(t, newCompany.Username, created.Username)
// }

func TestCompanyRepository_GetByID_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyRepo := postgresql.NewCompanyRepository(testDB)

	companyID := createTestCompany(t, ctx)

	retrieved, err := companyRepo.GetByID(ctx, companyID)

	assert.NoError(t, err)
	assert.Equal(t, companyID, retrieved.ID)
	assert.Equal(t, "Test Company", retrieved.Name)
}

func TestCompanyRepository_GetByUsername_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyRepo := postgresql.NewCompanyRepository(testDB)

	createTestCompany(t, ctx)

	retrieved, err := companyRepo.GetByUsername(ctx, "test-company")

	assert.NoError(t, err)
	assert.Equal(t, "test-company", retrieved.Username)
	assert.Equal(t, "Test Company", retrieved.Name)
}

func TestCompanyRepository_Update_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyRepo := postgresql.NewCompanyRepository(testDB)

	companyID := createTestCompany(t, ctx)

	updateReq := company.UpdateCompanyRequest{
		Name:    strPtr("Updated Company"),
		Address: strPtr("456 Oak Ave"),
	}

	err := companyRepo.Update(ctx, companyID, updateReq)

	assert.NoError(t, err)

	// Verify update
	updated, err := companyRepo.GetByID(ctx, companyID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Company", updated.Name)
	assert.NotNil(t, updated.Address)
	assert.Equal(t, "456 Oak Ave", *updated.Address)
}

func TestCompanyRepository_ExistsByID_Success(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyRepo := postgresql.NewCompanyRepository(testDB)

	companyID := createTestCompany(t, ctx)

	exists, err := companyRepo.ExistsByIDOrUsername(ctx, &companyID, nil)

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCompanyRepository_ExistsByID_NotFound(t *testing.T) {
	defer cleanupTestData(t)
	setupTestData(t)

	ctx := context.Background()
	companyRepo := postgresql.NewCompanyRepository(testDB)

	fakeBoolID := "00000000-0000-0000-0000-000000000000"

	exists, err := companyRepo.ExistsByIDOrUsername(ctx, &fakeBoolID, nil)

	assert.NoError(t, err)
	assert.False(t, exists)
}

// func TestCompanyRepository_List_Success(t *testing.T) {
// 	defer cleanupTestData(t)
// 	setupTestData(t)

// 	ctx := context.Background()
// 	companyRepo := postgresql.NewCompanyRepository(testDB)

// 	// Create multiple companies
// 	createTestCompany(t, ctx)
// 	createTestCompany(t, ctx)

// 	companies, err := companyRepo.List(ctx)

// 	assert.NoError(t, err)
// 	assert.GreaterOrEqual(t, len(companies), 2)
// }

// ===== HELPER FUNCTIONS =====

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
