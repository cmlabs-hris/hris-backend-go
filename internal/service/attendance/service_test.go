package attendance_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
	attendanceService "github.com/cmlabs-hris/hris-backend-go/internal/service/attendance"
	"github.com/go-chi/jwtauth/v5"
)

func ptr[T any](v T) *T {
	return &v
}

// testContext creates a context with JWT claims that the service can read.
func testContext(t *testing.T, employee *string, company *string) context.Context {
	t.Helper()

	if company == nil {
		company = ptr("comp-1")
	}
	if employee == nil {
		employee = ptr("emp-1")
	}

	ja := jwtauth.New("HS256", []byte("test-secret"), nil)
	token, _, err := ja.Encode(map[string]interface{}{
		"company_id":  *company,
		"employee_id": *employee,
	})
	if err != nil {
		t.Fatalf("failed to encode token: %v", err)
	}

	return jwtauth.NewContext(context.Background(), token, nil)
}

// testAttendanceRepo is a fake AttendanceRepository used in service tests.
// It embeds the real interface (so all methods exist) and overrides only
// the methods the test needs.
type testAttendanceRepo struct {
	attendance.AttendanceRepository
	getByIDFunc         func(ctx context.Context, id string, companyID string) (attendance.Attendance, error)
	deleteFunc          func(ctx context.Context, id string, companyID string) error
	listFunc            func(ctx context.Context, filter attendance.AttendanceFilter, companyID string) ([]attendance.Attendance, int64, error)
	updateFunc          func(ctx context.Context, attendance attendance.Attendance) error
	getMyAttendanceFunc func(ctx context.Context, employeeID string, filter attendance.MyAttendanceFilter, companyID string) ([]attendance.Attendance, int64, error)
}

func (far *testAttendanceRepo) GetByID(ctx context.Context, id string, companyID string) (attendance.Attendance, error) {
	return far.getByIDFunc(ctx, id, companyID)
}

func (far *testAttendanceRepo) Delete(ctx context.Context, id string, companyID string) error {
	return far.deleteFunc(ctx, id, companyID)
}

func (far *testAttendanceRepo) List(ctx context.Context, filter attendance.AttendanceFilter, companyID string) ([]attendance.Attendance, int64, error) {
	return far.listFunc(ctx, filter, companyID)
}

func (far *testAttendanceRepo) Update(ctx context.Context, attendance attendance.Attendance) error {
	return far.updateFunc(ctx, attendance)
}

func (far *testAttendanceRepo) GetMyAttendance(ctx context.Context, employeeID string, filter attendance.MyAttendanceFilter, companyID string) ([]attendance.Attendance, int64, error) {
	return far.getMyAttendanceFunc(ctx, employeeID, filter, companyID)
}

func validAttendance(id string) attendance.Attendance {
	return attendance.Attendance{
		ID:         id,
		EmployeeID: "emp-123",
		Date:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:     "waiting_approval",
		CompanyID:  "comp-123",
		CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func validUpdateAttendanceRequest(id string) attendance.UpdateAttendanceRequest {
	return attendance.UpdateAttendanceRequest{
		ID:                id,
		Date:              ptr("2026-01-01"),
		ClockInTime:       ptr("2026-01-01 12:01:02"),
		ClockOutTime:      ptr("2026-01-01 13:01:02"),
		Status:            ptr("present"),
		ClockInLatitude:   ptr(90.0),
		ClockInLongitude:  ptr(180.0),
		LateMinutes:       ptr(1),
		EarlyLeaveMinutes: ptr(5),
		OvertimeMinutes:   ptr(5),
	}
}

func TestListAttendance(t *testing.T) {
	tests := []struct {
		name      string
		repoTotal int
		repoAtts  []attendance.Attendance
		attFilter attendance.AttendanceFilter
		repoErr   error
		want      func(t *testing.T, err error, resp attendance.ListAttendanceResponse)
	}{
		{
			name:      "valid request",
			repoTotal: 20,
			repoAtts: []attendance.Attendance{
				validAttendance("1"),
				validAttendance("2"),
				validAttendance("3"),
				validAttendance("4"),
				validAttendance("5"),
			},
			attFilter: attendance.AttendanceFilter{
				Page:  3,
				Limit: 5,
			},
			repoErr: nil,
			want: func(t *testing.T, err error, resp attendance.ListAttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil err | got %v", err)
				}
				if resp.Showing != "11-15 of 20" {
					t.Errorf("want 11-15 of 20 | got %s", resp.Showing)
				}
				if len(resp.Attendances) != 5 {
					t.Errorf("Attendances length = %d, want 5", len(resp.Attendances))
				}
				if int(resp.TotalCount) != 20 {
					t.Errorf("want 20 total records, got %d instead", resp.TotalCount)
				}
				if len(resp.Attendances) > 0 && resp.Attendances[0].ID != "1" {
					t.Errorf("Attendances[0].ID = %q, want %q", resp.Attendances[0].ID, "1")
				}
				if resp.TotalPages != 4 {
					t.Errorf("want total of 4 pages, got %d instead", resp.TotalPages)
				}
				if resp.Limit != 5 {
					t.Errorf("want limit of 5 records, got %d instead", resp.Limit)
				}
				if resp.Page != 3 {
					t.Errorf("want current page to be 3, is instead %d", resp.Page)
				}
			},
		},
		{
			name:      "repo error",
			repoTotal: 10,
			repoAtts:  []attendance.Attendance{},
			attFilter: attendance.AttendanceFilter{},
			repoErr:   errors.New("db down"),
			want: func(t *testing.T, err error, resp attendance.ListAttendanceResponse) {
				t.Helper()
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "failed to list attendances") {
					t.Errorf("want error to contain 'failed to list attendances', but got %s", err.Error())
				}
			},
		},
		{
			name:      "attendance records is 0",
			repoTotal: 0,
			repoAtts:  []attendance.Attendance{},
			attFilter: attendance.AttendanceFilter{Page: 6, Limit: 5},
			repoErr:   nil,
			want: func(t *testing.T, err error, resp attendance.ListAttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil err | got %v", err)
				}
				if resp.TotalCount != 0 {
					t.Fatalf("want 0 attendance records, got %d instead", resp.TotalCount)
				}
				if resp.Showing != "0 of 0" {
					t.Errorf("want 0 of 0 showing, got %s instead", resp.Showing)
				}
				if len(resp.Attendances) != 0 {
					t.Errorf("Attendances length = %d, want 0", len(resp.Attendances))
				}
				if resp.Page != 6 {
					t.Errorf("want current page to be 6, is instead %d", resp.Page)
				}
				if resp.Limit != 5 {
					t.Errorf("want limit of 5 records, got %d instead", resp.Limit)
				}
			},
		},
		{
			name:      "show last attendance page that returns less than limit",
			repoTotal: 27,
			repoAtts: []attendance.Attendance{
				validAttendance("1"),
				validAttendance("2"),
			},
			attFilter: attendance.AttendanceFilter{Page: 6, Limit: 5},
			repoErr:   nil,
			want: func(t *testing.T, err error, resp attendance.ListAttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil err | got %v", err)
				}
				if resp.Showing != "26-27 of 27" {
					t.Errorf("want 26-27 of 27 | got %s", resp.Showing)
				}
				if len(resp.Attendances) != 2 {
					t.Errorf("Attendances length = %d, want 2", len(resp.Attendances))
				}
				if int(resp.TotalCount) != 27 {
					t.Errorf("want 27 total records, got %d instead", resp.TotalCount)
				}
				if len(resp.Attendances) > 0 && resp.Attendances[0].ID != "1" {
					t.Errorf("Attendances[0].ID = %q, want %q", resp.Attendances[0].ID, "1")
				}
				if resp.TotalPages != 6 {
					t.Errorf("want total of 6 pages, got %d instead", resp.TotalPages)
				}
				if resp.Limit != 5 {
					t.Errorf("want limit of 5 records, got %d instead", resp.Limit)
				}
				if resp.Page != 6 {
					t.Errorf("want current page to be 6, is instead %d", resp.Page)
				}
			},
		},
		{
			name:      "show every records",
			repoTotal: 5,
			repoAtts: []attendance.Attendance{
				validAttendance("1"),
				validAttendance("2"),
				validAttendance("3"),
				validAttendance("4"),
				validAttendance("5"),
			},
			attFilter: attendance.AttendanceFilter{Page: 1, Limit: 5},
			repoErr:   nil,
			want: func(t *testing.T, err error, resp attendance.ListAttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil err | got %v", err)
				}
				if resp.Showing != "1-5 of 5" {
					t.Errorf("want 1-5 of 5 | got %s", resp.Showing)
				}
				if len(resp.Attendances) != 5 {
					t.Errorf("Attendances length = %d, want 5", len(resp.Attendances))
				}
				if int(resp.TotalCount) != 5 {
					t.Errorf("want 5 total records, got %d instead", resp.TotalCount)
				}
				if len(resp.Attendances) > 0 && resp.Attendances[0].ID != "1" {
					t.Errorf("Attendances[0].ID = %q, want %q", resp.Attendances[0].ID, "1")
				}
				if resp.TotalPages != 1 {
					t.Errorf("want total of 1 pages, got %d instead", resp.TotalPages)
				}
				if resp.Limit != 5 {
					t.Errorf("want limit of 5 records, got %d instead", resp.Limit)
				}
				if resp.Page != 1 {
					t.Errorf("want current page to be 1, is instead %d", resp.Page)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attRepo := testAttendanceRepo{
				listFunc: func(ctx context.Context, filter attendance.AttendanceFilter, companyID string) ([]attendance.Attendance, int64, error) {
					return test.repoAtts, int64(test.repoTotal), test.repoErr
				},
			}
			srv := attendanceService.NewAttendanceService(attendanceService.Deps{
				AttendanceRepo: &attRepo,
			})
			ctx := testContext(t, nil, nil)
			resp, err := srv.ListAttendance(ctx, test.attFilter)
			test.want(t, err, resp)
		})
	}
}

func TestUpdateAttendance(t *testing.T) {
	tests := []struct {
		name           string
		request        attendance.UpdateAttendanceRequest
		repoAtt        attendance.Attendance
		repoErrGetByID error
		getByIDCall    int
		repoErrUpdate  error
		want           func(t *testing.T, err error, resp attendance.AttendanceResponse)
	}{
		{
			name:           "valid request",
			request:        validUpdateAttendanceRequest("1"),
			repoAtt:        validAttendance("1"),
			repoErrGetByID: nil,
			repoErrUpdate:  nil,
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if resp.ID != "1" {
					t.Errorf("want 1 ID but instead got %s", resp.ID)
				}
				if resp.Date != "2026-01-01" {
					t.Errorf("want 2026-01-01 but instead got %s", resp.Date)
				}
				if *resp.ClockInTime != "2026-01-01 12:01:02" {
					t.Errorf("want 2026-01-01 12:01:02 but instead got %s", *resp.ClockInTime)
				}
				if *resp.ClockOutTime != "2026-01-01 13:01:02" {
					t.Errorf("want 2026-01-01 13:01:02 but instead got %s", *resp.ClockOutTime)
				}
				if resp.Status != "present" {
					t.Errorf("want present but instead got %s", resp.Status)
				}
				if *resp.ClockInLatitude != 90.0 {
					t.Errorf("want 90.0 but instead got %f", *resp.ClockInLatitude)
				}
				if *resp.ClockInLongitude != 180.0 {
					t.Errorf("want 180.0 but instead got %f", *resp.ClockInLongitude)
				}
				if *resp.LateMinutes != 1 {
					t.Errorf("want 1 but instead got %d", *resp.LateMinutes)
				}
				if *resp.EarlyLeaveMinutes != 5 {
					t.Errorf("want 5 but instead got %d", *resp.EarlyLeaveMinutes)
				}
				if *resp.WorkingHours != 1.000000 {
					t.Errorf("want 1.000000 but instead got %f", *resp.WorkingHours)
				}
			},
		},
		{
			name: "work hours not recalculated (only clock in exists)",
			request: attendance.UpdateAttendanceRequest{
				ClockInTime: ptr("15:12:23"),
			},
			repoAtt: validAttendance("1"),
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if resp.WorkingHours != nil {
					t.Errorf("want working hours to be nil, but instead got %f", *resp.WorkingHours)
				}
			},
		},
		{
			name:           "repository GetByID error not found",
			request:        validUpdateAttendanceRequest("555"),
			repoErrGetByID: attendance.ErrAttendanceNotFound,
			getByIDCall:    1,
			repoErrUpdate:  nil,
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if !errors.Is(err, attendance.ErrAttendanceNotFound) {
					t.Errorf("want %v | got %v", attendance.ErrAttendanceNotFound, err)
				}
			},
		},
		{
			name:           "repository GetByID error db down",
			request:        validUpdateAttendanceRequest("412"),
			repoErrGetByID: errors.New("db down"),
			getByIDCall:    1,
			repoErrUpdate:  nil,
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if !strings.Contains(err.Error(), "failed to get attendance") {
					t.Errorf("want error to contain 'failed to get attendance' but instead got %v", err)
				}
			},
		},
		{
			name:           "repository GetByID error second call connection lost",
			request:        validUpdateAttendanceRequest("412"),
			repoAtt:        validAttendance("1"),
			repoErrGetByID: errors.New("connection lost"),
			getByIDCall:    2,
			repoErrUpdate:  nil,
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if !strings.Contains(err.Error(), "failed to get updated attendance") {
					t.Errorf("want error to contain 'failed to get updated attendance' but instead got %v", err)
				}
			},
		},
		{
			name:    "date field format invalid",
			repoAtt: validAttendance("1"),
			request: attendance.UpdateAttendanceRequest{
				Date: ptr("twenty six"),
			},
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if _, ok := errors.AsType[validator.ValidationErrors](err); !ok {
					t.Errorf("want %v | got %v", attendance.ErrDateFormat, err)
				}
			},
		},
		{
			name: "clock in format invalid",
			request: attendance.UpdateAttendanceRequest{
				ClockInTime: ptr("2006-01-02"),
			},
			repoAtt: validAttendance("1"),
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if _, ok := errors.AsType[validator.ValidationErrors](err); !ok {
					t.Errorf("want %v | got %v", attendance.ErrClockInFormat, err)
				}
			},
		},
		{
			name: "clock in valid full datetime",
			request: attendance.UpdateAttendanceRequest{
				ClockInTime: ptr("2006-01-02 12:12:12"),
			},
			repoAtt: validAttendance("1"),
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if *resp.ClockInTime != "2006-01-02 12:12:12" {
					t.Errorf("want '2006-01-02 12:12:12', but instead got %s", *resp.ClockInTime)
				}
			},
		},
		{
			name:    "clock in valid time only",
			repoAtt: validAttendance("1"),
			request: attendance.UpdateAttendanceRequest{
				ClockInTime: ptr("12:12:12"),
			},
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if *resp.ClockInTime != "2026-01-01 12:12:12" {
					t.Errorf("want '2026-01-01 12:12:12', but instead got %s", *resp.ClockInTime)
				}
			},
		},
		{
			name: "clock out format invalid",
			request: attendance.UpdateAttendanceRequest{
				ClockOutTime: ptr("2006-01-02"),
			},
			repoAtt: validAttendance("1"),
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if _, ok := errors.AsType[validator.ValidationErrors](err); !ok {
					t.Errorf("want %v | got %v", attendance.ErrClockOutFormat, err)
				}
			},
		},
		{
			name: "clock out valid full datetime",
			request: attendance.UpdateAttendanceRequest{
				ClockOutTime: ptr("2006-01-02 12:12:12"),
			},
			repoAtt: validAttendance("1"),
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if *resp.ClockOutTime != "2006-01-02 12:12:12" {
					t.Errorf("want '2006-01-02 12:12:12', but instead got %s", *resp.ClockOutTime)
				}
			},
		},
		{
			name:    "clock out valid time only",
			repoAtt: validAttendance("1"),
			request: attendance.UpdateAttendanceRequest{
				ClockOutTime: ptr("12:12:12"),
			},
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if *resp.ClockOutTime != "2026-01-01 12:12:12" {
					t.Errorf("want '2026-01-01 12:12:12', but instead got %s", *resp.ClockOutTime)
				}
			},
		},
		{
			name:          "repository Update error",
			request:       validUpdateAttendanceRequest("1"),
			repoAtt:       validAttendance("1"),
			repoErrUpdate: errors.New("db down"),
			want: func(t *testing.T, err error, resp attendance.AttendanceResponse) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if !strings.Contains(err.Error(), "failed to update attendance") {
					t.Errorf("want error to contain 'failed to update attendance' but instead got %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			existing := test.repoAtt
			callCount := 0
			attRepo := testAttendanceRepo{
				getByIDFunc: func(ctx context.Context, id, companyID string) (attendance.Attendance, error) {
					callCount++
					if callCount == test.getByIDCall {
						return existing, test.repoErrGetByID
					}
					return existing, nil
				},
				updateFunc: func(ctx context.Context, attendance attendance.Attendance) error {
					existing = attendance
					return test.repoErrUpdate
				},
			}
			srv := attendanceService.NewAttendanceService(attendanceService.Deps{
				AttendanceRepo: &attRepo,
			})
			ctx := testContext(t, nil, nil)
			resp, err := srv.UpdateAttendance(ctx, test.request)
			test.want(t, err, resp)
		})
	}
}

func TestDeleteAttendance(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		company *string
		want    func(t *testing.T, err error)
	}{
		{
			name:    "valid request",
			repoErr: nil,
			want: func(t *testing.T, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
			},
		},
		{
			name:    "repo delete error attendance not found",
			repoErr: attendance.ErrAttendanceNotFound,
			want: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, instead got nil")
				}
				if !errors.Is(err, attendance.ErrAttendanceNotFound) {
					t.Errorf("want %v, but instead got %v", attendance.ErrAttendanceNotFound, err)
				}
			},
		},
		{
			name:    "repo delete error db down",
			repoErr: errors.New("db down"),
			want: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, instead got nil")
				}
				if !strings.Contains(err.Error(), "failed to delete attendance") {
					t.Errorf("want error to contain 'failed to delete attendance', instead got %v", err.Error())
				}
			},
		},
		{
			name:    "missing company id",
			company: ptr(""),
			want: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, instead got nil")
				}
				if !strings.Contains(err.Error(), "company_id claim is missing or invalid") {
					t.Errorf("want error to contain 'company_id claim is missing or invalid', instead got %v", err.Error())
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attRepo := testAttendanceRepo{
				deleteFunc: func(ctx context.Context, id, companyID string) error {
					return test.repoErr
				},
			}
			srv := attendanceService.NewAttendanceService(attendanceService.Deps{
				AttendanceRepo: &attRepo,
			})
			err := srv.DeleteAttendance(testContext(t, nil, test.company), "1")
			test.want(t, err)
		})
	}
}

func TestGetMyAttendance(t *testing.T) {
	tests := []struct {
		name      string
		filter    attendance.MyAttendanceFilter
		repoAtts  []attendance.Attendance
		repoTotal int
		employee  *string
		company   *string
		repoErr   error
		want      func(t *testing.T, resp attendance.ListAttendanceResponse, err error)
	}{
		{
			name: "valid request",
			filter: attendance.MyAttendanceFilter{
				Page:  2,
				Limit: 5,
			},
			repoAtts: []attendance.Attendance{
				validAttendance("1"),
				validAttendance("2"),
				validAttendance("3"),
				validAttendance("4"),
				validAttendance("5"),
			},
			repoTotal: 20,
			want: func(t *testing.T, resp attendance.ListAttendanceResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if resp.TotalCount != 20 {
					t.Errorf("want 20 total, but instead got %d", resp.TotalCount)
				}
				if len(resp.Attendances) != 5 {
					t.Errorf("want 5 attendances amount, but instead got %d", len(resp.Attendances))
				}
				if resp.TotalPages != 4 {
					t.Errorf("want 4 total pages, but instead have %d", resp.TotalPages)
				}
				if resp.Showing != "6-10 of 20" {
					t.Errorf("want '6-10 of 20', but instead show %s", resp.Showing)
				}
			},
		},
		{
			name:    "company_id missing",
			company: ptr(""),
			filter: attendance.MyAttendanceFilter{
				Page:  2,
				Limit: 5,
			},
			want: func(t *testing.T, resp attendance.ListAttendanceResponse, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if !strings.Contains(err.Error(), "company_id claim is missing or invalid") {
					t.Errorf("want error to contain 'company_id claim is missing or invalid', but instead got %s", err.Error())
				}
			},
		},
		{
			name:     "employee_id missing",
			employee: ptr(""),
			want: func(t *testing.T, resp attendance.ListAttendanceResponse, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if !strings.Contains(err.Error(), "employee_id claim is missing or invalid") {
					t.Errorf("want error to contain 'employee_id claim is missing or invalid', but instead got %s", err.Error())
				}
			},
		},
		{
			name: "repository GetMyAttendance error db down",
			filter: attendance.MyAttendanceFilter{
				Page:  2,
				Limit: 5,
			},
			repoErr: errors.New("db down"),
			want: func(t *testing.T, resp attendance.ListAttendanceResponse, err error) {
				t.Helper()
				if err == nil {
					t.Fatalf("want error, but got nil instead")
				}
				if !strings.Contains(err.Error(), "failed to get my attendance") {
					t.Errorf("want error to contain 'failed to get my attendance', but instead got %s", err.Error())
				}
			},
		},
		{
			name: "zero attendances",
			filter: attendance.MyAttendanceFilter{
				Page:  2,
				Limit: 5,
			},
			repoAtts:  []attendance.Attendance{},
			repoTotal: 0,
			want: func(t *testing.T, resp attendance.ListAttendanceResponse, err error) {
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if resp.Showing != "0 of 0" {
					t.Errorf("want '0 of 0', but instead show %s", resp.Showing)
				}
				if resp.TotalCount != 0 {
					t.Errorf("want 0 total, but instead got %d", resp.TotalCount)
				}
				if len(resp.Attendances) != 0 {
					t.Errorf("want 0 attendances amount, but instead got %d", len(resp.Attendances))
				}
				if resp.TotalPages != 0 {
					t.Errorf("want 0 total pages, but instead have %d", resp.TotalPages)
				}
			},
		},
		{
			name: "showing full attendance records",
			filter: attendance.MyAttendanceFilter{
				Page:  1,
				Limit: 5,
			},
			repoAtts: []attendance.Attendance{
				validAttendance("1"),
				validAttendance("2"),
				validAttendance("3"),
				validAttendance("4"),
				validAttendance("5"),
			},
			repoTotal: 5,
			want: func(t *testing.T, resp attendance.ListAttendanceResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if resp.TotalCount != 5 {
					t.Errorf("want 5 total, but instead got %d", resp.TotalCount)
				}
				if len(resp.Attendances) != 5 {
					t.Errorf("want 5 attendances amount, but instead got %d", len(resp.Attendances))
				}
				if resp.TotalPages != 1 {
					t.Errorf("want 1 total pages, but instead have %d", resp.TotalPages)
				}
				if resp.Showing != "1-5 of 5" {
					t.Errorf("want '1-5 of 5', but instead show %s", resp.Showing)
				}
			},
		},
		{
			name: "showing last page attendance records",
			filter: attendance.MyAttendanceFilter{
				Page:  2,
				Limit: 5,
			},
			repoAtts: []attendance.Attendance{
				validAttendance("1"),
				validAttendance("2"),
			},
			repoTotal: 7,
			want: func(t *testing.T, resp attendance.ListAttendanceResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("want empty error, instead got %v", err)
				}
				if resp.TotalCount != 7 {
					t.Errorf("want 7 total, but instead got %d", resp.TotalCount)
				}
				if len(resp.Attendances) != 2 {
					t.Errorf("want 2 attendances amount, but instead got %d", len(resp.Attendances))
				}
				if resp.TotalPages != 2 {
					t.Errorf("want 1 total pages, but instead have %d", resp.TotalPages)
				}
				if resp.Showing != "6-7 of 7" {
					t.Errorf("want '6-7 of 7', but instead show %s", resp.Showing)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attRepo := testAttendanceRepo{
				getMyAttendanceFunc: func(ctx context.Context, employeeID string, filter attendance.MyAttendanceFilter, companyID string) ([]attendance.Attendance, int64, error) {
					return test.repoAtts, int64(test.repoTotal), test.repoErr
				},
			}
			srv := attendanceService.NewAttendanceService(attendanceService.Deps{
				AttendanceRepo: &attRepo,
			})
			ctx := testContext(t, test.employee, test.company)
			resp, err := srv.GetMyAttendance(ctx, test.filter)
			test.want(t, resp, err)
		})
	}
}
