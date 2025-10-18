# Repository Integration Tests

Comprehensive integration tests untuk PostgreSQL repository layer, mengakses database secara langsung tanpa mock.

## Prerequisites

1. **PostgreSQL** sudah installed dan running
2. **Create test database:**
   ```bash
   createdb hris_test
   ```

3. **Run migrations** untuk setup schema di test database:
   ```bash
   # Copy migration files ke test database
   psql -h localhost -U postgres -d hris_test < internal/infrastructure/database/postgresql/migrations/000001_initial_schema.up.sql
   ```

## Environment Setup

Set environment variable sebelum menjalankan test:

```bash
# Linux/Mac
export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/hris_test?sslmode=disable"

# Windows PowerShell
$env:TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/hris_test?sslmode=disable"

# Windows CMD
set TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/hris_test?sslmode=disable
```

Atau gunakan `.env` file dan load dengan tool seperti `godotenv`.

## Running Tests

### Run semua tests di repository package:
```bash
go test -v ./internal/repository/postgresql/...
```

### Run specific test file:
```bash
go test -v ./internal/repository/postgresql/ -run TestUserRepository
```

### Run specific test:
```bash
go test -v ./internal/repository/postgresql/ -run TestUserRepository_Create_Success
```

### Run dengan timeout:
```bash
go test -v -timeout 30s ./internal/repository/postgresql/...
```

### Run dengan race detector:
```bash
go test -race ./internal/repository/postgresql/...
```

### Run dengan coverage:
```bash
go test -cover ./internal/repository/postgresql/...
go test -coverprofile=coverage.out ./internal/repository/postgresql/...
go tool cover -html=coverage.out
```

## Test Coverage

### User Repository Tests
- ✅ Create user successfully
- ✅ Get user by email (success & not found)
- ✅ Get user by ID
- ✅ Link Google account
- ✅ Link password account
- ✅ Update user
- ✅ Delete user

### Company Repository Tests
- ✅ Create company successfully
- ✅ Get company by ID
- ✅ Get company by username
- ✅ Update company
- ✅ Check if company exists by ID
- ✅ List all companies

## Test Structure

Setiap test function mengikuti pattern:

1. **Setup** - Prepare test data
2. **Act** - Execute repository method
3. **Assert** - Verify results

```go
func TestUserRepository_Create_Success(t *testing.T) {
    // Setup
    defer cleanupTestData(t)
    setupTestData(t)
    
    ctx := context.Background()
    companyID := createTestCompany(t, ctx)
    userRepo := postgresql.NewUserRepository(testDB)
    
    // Act
    created, err := userRepo.Create(ctx, newUser)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, created.ID)
}
```

## Helper Functions

- `setupTestData(t)` - Truncate tables sebelum test
- `cleanupTestData(t)` - Truncate tables setelah test
- `createTestCompany(t, ctx)` - Buat company untuk testing
- `createTestUser(t, ctx, companyID)` - Buat user untuk testing
- `strPtr(s)` - Konversi string ke pointer
- `boolPtr(b)` - Konversi bool ke pointer

## Best Practices

1. **Always cleanup**: Gunakan `defer cleanupTestData(t)` setelah setup
2. **Use context**: Pass context ke semua database operations
3. **Verify transactions**: Pastikan data berhasil disimpan dengan query ulang
4. **Test edge cases**: Test not found, duplicate, invalid data
5. **Use meaningful names**: Test function names harus jelas apa yang ditest

## Troubleshooting

### Connection Refused
```
connection refused
```
Pastikan PostgreSQL running di localhost:5432

### Database Doesn't Exist
```
FATAL: database "hris_test" does not exist
```
Run: `createdb hris_test`

### Schema Not Found
```
ERROR: relation "users" does not exist
```
Run migrations ke test database

### Timeout
Jika test timeout, increase dengan `-timeout`:
```bash
go test -timeout 60s ./internal/repository/postgresql/...
```

## CI/CD Integration

Untuk GitHub Actions atau CI/CD lainnya:

```yaml
- name: Run Repository Tests
  env:
    TEST_DATABASE_URL: postgres://postgres:postgres@localhost:5432/hris_test?sslmode=disable
  run: |
    go test -v -race -coverprofile=coverage.out ./internal/repository/postgresql/...
```

Make sure PostgreSQL service di-setup di CI/CD pipeline.
