# Service Integration Tests - Created Files

## Overview
Created comprehensive integration tests for the service layer that use the test database directly (no mocks), following the same pattern as the repository tests in `#file:postgresql_test`.

## Files Created

### 1. `internal/service/auth/auth_service_test.go`
**Purpose:** Integration tests for AuthService with real database access

**Test Functions:**
- `TestAuthService_Login_Success` - Test successful login with valid credentials
- `TestAuthService_Login_InvalidPassword` - Test login rejection with wrong password
- `TestAuthService_Login_UserNotFound` - Test login rejection when user doesn't exist
- `TestAuthService_LoginWithGoogle_NewUser` - Test OAuth Google login creating new user
- `TestAuthService_LoginWithGoogle_ExistingUser` - Test OAuth Google login with existing user
- `TestAuthService_RevokeRefreshToken_Success` - Test token revocation for logout

**Key Features:**
- Uses `authTestInit()` for lazy database connection
- `truncateAuthTables()` cleans up test data between tests
- Helper functions: `createAuthTestCompany()`, `createAuthTestUser()`
- Parallel test execution with `t.Parallel()`
- Tests both success and error paths
- Direct database interaction without mocks

**Test Flow:**
1. Initialize database connection (once)
2. Setup test data (company, user)
3. Create service with real repositories
4. Execute service method
5. Assert results and verify database state
6. Cleanup tables after test

### 2. `internal/service/company/company_service_test.go`
**Purpose:** Integration tests for CompanyService with real database access

**Test Functions:**
- `TestCompanyService_Create_Success` - Test creating a new company
- `TestCompanyService_GetByID_Success` - Test retrieving company by ID
- `TestCompanyService_GetByID_NotFound` - Test retrieval of non-existent company
- `TestCompanyService_Update_Success` - Test updating company fields
- `TestCompanyService_Update_PartialFields` - Test partial updates
- `TestCompanyService_Update_NoFields` - Test error when no fields to update
- `TestCompanyService_Delete_Success` - Test company deletion
- `TestCompanyService_List_Success` - Test listing all companies

**Key Features:**
- Uses `companyTestInit()` for lazy database connection
- `truncateCompanyTables()` cleans up test data between tests
- Helper function: `createCompanyTestCompany()`
- Parallel test execution with `t.Parallel()`
- Tests CRUD operations and edge cases
- Verifies database state after operations
- Direct database interaction without mocks

## Test Setup Configuration

### Environment Variables
```powershell
# Windows PowerShell
$env:TEST_DATABASE_URL="postgres://postgres:root@localhost:5432/cmlabs_hris_test?sslmode=disable"
```

### Database Prerequisites
```bash
# Create test database
createdb cmlabs_hris_test

# Run migrations
migrate -path internal/infrastructure/database/postgresql/migrations \
        -database $TEST_DATABASE_URL up
```

## Running the Tests

### Run all service tests
```bash
go test -v ./internal/service/...
```

### Run specific service tests
```bash
# Auth service tests only
go test -v ./internal/service/auth/...

# Company service tests only
go test -v ./internal/service/company/...
```

### Run with coverage
```bash
go test -v -cover ./internal/service/auth/...
go test -v -cover ./internal/service/company/...
```

### Run with race detector
```bash
go test -race -v ./internal/service/...
```

## Shared Test Patterns

All service tests follow the same pattern established in `#file:postgresql_test`:

1. **Lazy Initialization**
   ```go
   func serviceTestInit() {
       if testDB != nil { return }
       // connect to database
   }
   ```

2. **Per-Test Cleanup**
   ```go
   func TestExample(t *testing.T) {
       defer truncateTables(t, ctx)
       // test code
   }
   ```

3. **Helper Functions**
   ```go
   companyID := createTestCompany(t, ctx)
   user := createTestUser(t, ctx, companyID)
   ```

4. **Parallel Execution**
   ```go
   func TestExample(t *testing.T) {
       t.Parallel()
       // tests run concurrently
   }
   ```

## Key Differences from Repository Tests

**Similarities:**
- Direct database access (no mocks)
- Same truncation/cleanup pattern
- Parallel test execution
- Helper functions for test data creation

**Differences:**
- Service layer tests create real service instances with injected repositories
- Services test business logic, not just CRUD operations
- Services may coordinate between multiple repositories
- Error handling at service layer vs repository layer
- Transaction handling and constraints validation

## Test Execution Notes

### Known Issues to Resolve
1. Migrations must be run on test database before tests can pass
2. Some tests use hardcoded company usernames which may cause constraint violations with parallel execution
3. Tests should use unique identifiers to avoid conflicts

### Future Enhancements
1. Add test data factories with unique identifiers
2. Add service error handling tests
3. Add transaction boundary tests
4. Add authorization/permission tests for company operations
5. Add service logging verification

## Integration with CI/CD

These tests can be run in GitHub Actions:
```yaml
- name: Run Service Integration Tests
  env:
    TEST_DATABASE_URL: postgres://postgres:postgres@localhost:5432/cmlabs_hris_test?sslmode=disable
  run: |
    go test -v -race ./internal/service/...
```

## Next Steps

1. **Run migrations on test database** - Required before tests pass
2. **Fix company username uniqueness** - Use random identifiers or UUIDs
3. **Add employee service tests** - Following same pattern
4. **Add additional edge case tests** - Based on business logic requirements
5. **Add authorization tests** - Admin vs regular user permissions
6. **Add transaction tests** - Verify ACID properties at service level

## Related Files
- Repository integration tests: `#file:postgresql_test`
- Auth domain: `internal/domain/auth/`
- Company domain: `internal/domain/company/`
- Auth service implementation: `internal/service/auth/service.go`
- Company service implementation: `internal/service/company/service.go`
