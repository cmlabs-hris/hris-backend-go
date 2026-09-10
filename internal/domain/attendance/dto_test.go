package attendance_test

import (
	"errors"
	"mime/multipart"
	"testing"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

func ptr[T any](v T) *T {
	return &v
}

func assertValidationErrors(t *testing.T, err error, wantErrFields []string) {
	t.Helper()

	if len(wantErrFields) == 0 {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var errVal validator.ValidationErrors
	if !errors.As(err, &errVal) {
		t.Fatalf("error type %v, want type ValidationError", err)
	}

	errorFields := errVal.ToMap()
	if len(wantErrFields) != len(errorFields) {
		t.Errorf("want %d errors but found %d", len(wantErrFields), len(errorFields))
	}

	for _, v := range wantErrFields {
		if _, ok := errorFields[v]; !ok {
			t.Errorf("want error field %s, but not found", v)
		}
	}
}

func validClockInRequest() attendance.ClockInRequest {

	return attendance.ClockInRequest{
		EmployeeID: "1231321312",
		Latitude:   ptr(70.0),
		Longitude:  ptr(150.0),
		FileHeader: &multipart.FileHeader{
			Filename: "clockin.jpg",
			Size:     5,
		},
	}
}

func makeClockInRequest(modify func(*attendance.ClockInRequest)) attendance.ClockInRequest {
	req := validClockInRequest()
	if modify != nil {
		modify(&req)
	}
	return req
}

func TestClockInRequest(t *testing.T) {
	requests := []struct {
		name          string
		request       attendance.ClockInRequest
		wantErrFields []string
	}{
		{
			name:          "valid request",
			request:       validClockInRequest(),
			wantErrFields: nil,
		},
		{
			name: "missing employee data",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.EmployeeID = ""
			}),
			wantErrFields: []string{"employee_id"},
		},
		{
			name: "missing employee data whitespace",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.EmployeeID = " "
			}),
			wantErrFields: []string{"employee_id"},
		},
		{
			name: "missing latitude field",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Latitude = nil
			}),
			wantErrFields: []string{"latitude"},
		},
		{
			name: "latitude less than -90",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Latitude = ptr(-95.0)
			}),
			wantErrFields: []string{"latitude"},
		},
		{
			name: "latitude exactly -90",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Latitude = ptr(-90.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "latitude exactly 90",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Latitude = ptr(90.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "latitude more than 90",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Latitude = ptr(91.0)
			}),
			wantErrFields: []string{"latitude"},
		},
		{
			name: "missing longitude field",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Longitude = nil
			}),
			wantErrFields: []string{"longitude"},
		},
		{
			name: "longitude less than -180",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Longitude = ptr(-185.0)
			}),
			wantErrFields: []string{"longitude"},
		},
		{
			name: "longitude exactly -180",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Longitude = ptr(-180.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "longitude exactly 180",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Longitude = ptr(180.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "longitude more than 180",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Longitude = ptr(181.0)
			}),
			wantErrFields: []string{"longitude"},
		},
		{
			name: "missing file field",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = nil
			}),
			wantErrFields: []string{"file"},
		},
		{
			name: "missing file extension",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma",
					Size:     5,
				}
			}),
			wantErrFields: []string{"file"},
		},
		{
			name: "uppercase file extension",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.JPG",
					Size:     5,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "jpg file extension",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpg",
					Size:     5,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "jpeg file extension",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpeg",
					Size:     5,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "png file extension",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.png",
					Size:     5,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "invalid file type extension",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.gif",
					Size:     5,
				}
			}),
			wantErrFields: []string{"file"},
		},
		{
			name: "file size exceeds 10MB",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpg",
					Size:     10<<20 + 1,
				}
			}),
			wantErrFields: []string{"file"},
		},
		{
			name: "file size exactly 10MB",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpg",
					Size:     10 << 20,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "invalid latitude, longitude, and file size exceeds 10MB",
			request: makeClockInRequest(func(cir *attendance.ClockInRequest) {
				cir.Latitude = ptr(-95.0)
				cir.Longitude = ptr(-185.0)
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpg",
					Size:     10<<20 + 1,
				}
			}),
			wantErrFields: []string{"latitude", "longitude", "file"},
		},
	}

	for _, request := range requests {
		t.Run(request.name, func(t *testing.T) {
			err := request.request.Validate()

			assertValidationErrors(t, err, request.wantErrFields)
		})
	}
}

func validClockOutRequest() attendance.ClockOutRequest {

	return attendance.ClockOutRequest{
		EmployeeID: "1231321312",
		Latitude:   ptr(70.0),
		Longitude:  ptr(150.0),
		FileHeader: &multipart.FileHeader{
			Filename: "clockout.jpg",
			Size:     5,
		},
	}
}

func makeClockOutRequest(modify func(*attendance.ClockOutRequest)) attendance.ClockOutRequest {
	req := validClockOutRequest()
	if modify != nil {
		modify(&req)
	}
	return req
}

func TestClockOutRequest(t *testing.T) {
	requests := []struct {
		name          string
		request       attendance.ClockOutRequest
		wantErrFields []string
	}{
		{
			name:          "valid request",
			request:       validClockOutRequest(),
			wantErrFields: nil,
		},
		{
			name: "missing employee data",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.EmployeeID = ""
			}),
			wantErrFields: []string{"employee_id"},
		},
		{
			name: "missing employee data whitespace",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.EmployeeID = " "
			}),
			wantErrFields: []string{"employee_id"},
		},
		{
			name: "missing latitude field",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Latitude = nil
			}),
			wantErrFields: []string{"latitude"},
		},
		{
			name: "latitude less than -90",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Latitude = ptr(-95.0)
			}),
			wantErrFields: []string{"latitude"},
		},
		{
			name: "latitude exactly -90",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Latitude = ptr(-90.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "latitude exactly 90",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Latitude = ptr(90.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "latitude more than 90",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Latitude = ptr(91.0)
			}),
			wantErrFields: []string{"latitude"},
		},
		{
			name: "missing longitude field",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Longitude = nil
			}),
			wantErrFields: []string{"longitude"},
		},
		{
			name: "longitude less than -180",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Longitude = ptr(-185.0)
			}),
			wantErrFields: []string{"longitude"},
		},
		{
			name: "longitude exactly -180",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Longitude = ptr(-180.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "longitude exactly 180",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Longitude = ptr(180.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "longitude more than 180",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Longitude = ptr(181.0)
			}),
			wantErrFields: []string{"longitude"},
		},
		{
			name: "missing file field",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = nil
			}),
			wantErrFields: []string{"file"},
		},
		{
			name: "missing file extension",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma",
					Size:     5,
				}
			}),
			wantErrFields: []string{"file"},
		},
		{
			name: "uppercase file extension",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.JPG",
					Size:     5,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "jpg file extension",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpg",
					Size:     5,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "jpeg file extension",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpeg",
					Size:     5,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "png file extension",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.png",
					Size:     5,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "invalid file type extension",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.gif",
					Size:     5,
				}
			}),
			wantErrFields: []string{"file"},
		},
		{
			name: "file size exceeds 10MB",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpg",
					Size:     10<<20 + 1,
				}
			}),
			wantErrFields: []string{"file"},
		},
		{
			name: "file size exactly 10MB",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpg",
					Size:     10 << 20,
				}
			}),
			wantErrFields: nil,
		},
		{
			name: "invalid latitude, longitude, and file size exceeds 10MB",
			request: makeClockOutRequest(func(cir *attendance.ClockOutRequest) {
				cir.Latitude = ptr(-95.0)
				cir.Longitude = ptr(-185.0)
				cir.FileHeader = &multipart.FileHeader{
					Filename: "sigma.jpg",
					Size:     10<<20 + 1,
				}
			}),
			wantErrFields: []string{"latitude", "longitude", "file"},
		},
	}

	for _, request := range requests {
		t.Run(request.name, func(t *testing.T) {
			err := request.request.Validate()
			assertValidationErrors(t, err, request.wantErrFields)
		})
	}
}

func validAttendanceFilter() attendance.AttendanceFilter {
	return attendance.AttendanceFilter{
		EmployeeID:   ptr("123"),
		EmployeeName: ptr("jonathan"),
		Status:       ptr("present"),
		SortBy:       "date",
		SortOrder:    "desc",
		Date:         nil,
		StartDate:    nil,
		EndDate:      nil,
	}
}

func makeAttendanceFilter(modify func(*attendance.AttendanceFilter)) attendance.AttendanceFilter {
	req := validAttendanceFilter()
	if modify != nil {
		modify(&req)
	}
	return req
}

func TestAttendanceFilter(t *testing.T) {
	requests := []struct {
		name          string
		request       attendance.AttendanceFilter
		wantErrFields []string
		wantPage      int
		wantLimit     int
		wantSortBy    string
		wantSortOrder string
	}{
		{
			name:          "valid request",
			request:       validAttendanceFilter(),
			wantErrFields: nil,
			wantPage:      1,
			wantLimit:     20,
		},
		{
			name:          "page value less than 0",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Page = -5 }),
			wantErrFields: []string{"page"},
		},
		{
			name:          "page exactly 0 and use default value",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Page = 0 }),
			wantErrFields: nil,
			wantPage:      1,
		},
		{
			name:          "limit value less than 0",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Limit = -5 }),
			wantErrFields: []string{"limit"},
		},
		{
			name:          "limit value more than 100",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Limit = 500 }),
			wantErrFields: []string{"limit"},
		},
		{
			name:          "limit exactly 0 and use default value",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Limit = 0 }),
			wantErrFields: nil,
			wantLimit:     20,
		},
		{
			name:          "limit exactly 100",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Limit = 100 }),
			wantErrFields: nil,
		},
		{
			name:          "status value is not one of the allowed",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { *af.Status = "sigma" }),
			wantErrFields: []string{"status"},
		},
		{
			name:          "allowed status value is typo",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { *af.Status = "presentt" }),
			wantErrFields: []string{"status"},
		},
		{
			name:          "status value is one of the allowed",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { *af.Status = "absent" }),
			wantErrFields: nil,
		},
		{
			name:          "status is nil skips validation",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Status = nil }),
			wantErrFields: nil,
		},
		{
			name:          "date format is not YYYY-MM-DD",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Date = ptr("months") }),
			wantErrFields: []string{"date"},
		},
		{
			name:          "date format is valid YYYY-MM-DD",
			request:       makeAttendanceFilter(func(af *attendance.AttendanceFilter) { af.Date = ptr("2026-01-01") }),
			wantErrFields: nil,
		},
		{
			name: "date has exact date and start date / end date",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.Date = ptr("2026-01-01")
				af.StartDate = ptr("2026-05-05")
				af.EndDate = ptr("2026-07-07")
			}),
			wantErrFields: []string{"date"},
		},
		{
			name: "start_date format is not valid YYYY-MM-DD",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.StartDate = ptr("months")
			}),
			wantErrFields: []string{"start_date"},
		},
		{
			name: "start_date is valid YYYY-MM-DD",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.StartDate = ptr("2026-08-10")
			}),
			wantErrFields: nil,
		},
		{
			name: "end_date format is not valid YYYY-MM-DD",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.EndDate = ptr("months")
			}),
			wantErrFields: []string{"end_date"},
		},
		{
			name: "end_date is valid YYYY-MM-DD",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.EndDate = ptr("2026-08-10")
			}),
			wantErrFields: nil,
		},
		{
			name: "start_date valid before end_date",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.StartDate = ptr("2026-05-05")
				af.EndDate = ptr("2026-08-10")
			}),
			wantErrFields: nil,
		},
		{
			name: "start_date not valid before end_date",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.StartDate = ptr("2026-08-10")
				af.EndDate = ptr("2026-05-05")
			}),
			wantErrFields: []string{"start_date"},
		},
		{
			name: "sort_by is not one of the allowed",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.SortBy = "company"
			}),
			wantErrFields: []string{"sort_by"},
		},
		{
			name: "sort_by allowed is typo",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.SortBy = "datee"
			}),
			wantErrFields: []string{"sort_by"},
		},
		{
			name: "sort_by is valid one of the allowed",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.SortBy = "employee_name"
			}),
			wantErrFields: nil,
		},
		{
			name: "sort_by using default value",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.SortBy = ""
			}),
			wantErrFields: nil,
			wantSortBy:    "date",
		},

		{
			name: "sort_order is not one of the allowed",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.SortOrder = "middle"
			}),
			wantErrFields: []string{"sort_order"},
		},
		{
			name: "sort_order allowed is typo",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.SortOrder = "ascc"
			}),
			wantErrFields: []string{"sort_order"},
		},
		{
			name: "sort_order is valid one of the allowed",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.SortOrder = "asc"
			}),
			wantErrFields: nil,
		},
		{
			name: "sort_order using default value",
			request: makeAttendanceFilter(func(af *attendance.AttendanceFilter) {
				af.SortOrder = ""
			}),
			wantErrFields: nil,
			wantSortOrder: "desc",
		},
	}

	for _, request := range requests {
		t.Run(request.name, func(t *testing.T) {
			err := request.request.Validate()
			assertValidationErrors(t, err, request.wantErrFields)

			if request.wantPage != 0 && request.request.Page != request.wantPage {
				t.Errorf("Page = %d, want %d", request.request.Page, request.wantPage)
			}
			if request.wantLimit != 0 && request.request.Limit != request.wantLimit {
				t.Errorf("Limit = %d, want %d", request.request.Limit, request.wantLimit)
			}
			if request.wantSortBy != "" && request.request.SortBy != request.wantSortBy {
				t.Errorf("SortBy = %q, want %q", request.request.SortBy, request.wantSortBy)
			}
			if request.wantSortOrder != "" && request.request.SortOrder != request.wantSortOrder {
				t.Errorf("SortOrder = %q, want %q", request.request.SortOrder, request.wantSortOrder)
			}
		})
	}
}

func validMyAttendanceFilter() attendance.MyAttendanceFilter {
	return attendance.MyAttendanceFilter{
		Status:    ptr("present"),
		SortBy:    "date",
		SortOrder: "desc",
		Date:      nil,
		StartDate: nil,
		EndDate:   nil,
	}
}

func makeMyAttendanceFilter(modify func(*attendance.MyAttendanceFilter)) attendance.MyAttendanceFilter {
	req := validMyAttendanceFilter()
	if modify != nil {
		modify(&req)
	}
	return req
}

func TestMyAttendanceFilter(t *testing.T) {
	requests := []struct {
		name          string
		request       attendance.MyAttendanceFilter
		wantErrFields []string
		wantPage      int
		wantLimit     int
		wantSortBy    string
		wantSortOrder string
	}{
		{
			name:          "valid request",
			request:       validMyAttendanceFilter(),
			wantErrFields: nil,
			wantPage:      1,
			wantLimit:     20,
		},
		{
			name:          "page value less than 0",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Page = -5 }),
			wantErrFields: []string{"page"},
		},
		{
			name:          "page exactly 0 and use default value",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Page = 0 }),
			wantErrFields: nil,
			wantPage:      1,
		},
		{
			name:          "limit value less than 0",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Limit = -5 }),
			wantErrFields: []string{"limit"},
		},
		{
			name:          "limit value more than 100",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Limit = 500 }),
			wantErrFields: []string{"limit"},
		},
		{
			name:          "limit exactly 0 and use default value",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Limit = 0 }),
			wantErrFields: nil,
			wantLimit:     20,
		},
		{
			name:          "limit exactly 100",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Limit = 100 }),
			wantErrFields: nil,
		},
		{
			name:          "status value is not one of the allowed",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { *af.Status = "sigma" }),
			wantErrFields: []string{"status"},
		},
		{
			name:          "allowed status value is typo",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { *af.Status = "presentt" }),
			wantErrFields: []string{"status"},
		},
		{
			name:          "status value is one of the allowed",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { *af.Status = "absent" }),
			wantErrFields: nil,
		},
		{
			name:          "status is nil skips validation",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Status = nil }),
			wantErrFields: nil,
		},
		{
			name:          "waiting_approval status is allowed",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { *af.Status = "waiting_approval" }),
			wantErrFields: nil,
		},
		{
			name:          "date format is not YYYY-MM-DD",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Date = ptr("months") }),
			wantErrFields: []string{"date"},
		},
		{
			name:          "date format is valid YYYY-MM-DD",
			request:       makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) { af.Date = ptr("2026-01-01") }),
			wantErrFields: nil,
		},
		{
			name: "start_date format is not valid YYYY-MM-DD",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.StartDate = ptr("months")
			}),
			wantErrFields: []string{"start_date"},
		},
		{
			name: "start_date is valid YYYY-MM-DD",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.StartDate = ptr("2026-08-10")
			}),
			wantErrFields: nil,
		},
		{
			name: "end_date format is not valid YYYY-MM-DD",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.EndDate = ptr("months")
			}),
			wantErrFields: []string{"end_date"},
		},
		{
			name: "end_date is valid YYYY-MM-DD",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.EndDate = ptr("2026-08-10")
			}),
			wantErrFields: nil,
		},
		{
			name: "sort_by is not one of the allowed",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.SortBy = "company"
			}),
			wantErrFields: []string{"sort_by"},
		},
		{
			name: "sort_by is valid one of the allowed",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.SortBy = "clock_in_time"
			}),
			wantErrFields: nil,
		},
		{
			name: "sort_by using default value",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.SortBy = ""
			}),
			wantErrFields: nil,
			wantSortBy:    "date",
		},
		{
			name: "sort_order is not one of the allowed",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.SortOrder = "middle"
			}),
			wantErrFields: []string{"sort_order"},
		},
		{
			name: "sort_order is valid one of the allowed",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.SortOrder = "asc"
			}),
			wantErrFields: nil,
		},
		{
			name: "sort_order using default value",
			request: makeMyAttendanceFilter(func(af *attendance.MyAttendanceFilter) {
				af.SortOrder = ""
			}),
			wantErrFields: nil,
			wantSortOrder: "desc",
		},
	}

	for _, request := range requests {
		t.Run(request.name, func(t *testing.T) {
			err := request.request.Validate()
			assertValidationErrors(t, err, request.wantErrFields)

			if request.wantPage != 0 && request.request.Page != request.wantPage {
				t.Errorf("Page = %d, want %d", request.request.Page, request.wantPage)
			}
			if request.wantLimit != 0 && request.request.Limit != request.wantLimit {
				t.Errorf("Limit = %d, want %d", request.request.Limit, request.wantLimit)
			}
			if request.wantSortBy != "" && request.request.SortBy != request.wantSortBy {
				t.Errorf("SortBy = %q, want %q", request.request.SortBy, request.wantSortBy)
			}
			if request.wantSortOrder != "" && request.request.SortOrder != request.wantSortOrder {
				t.Errorf("SortOrder = %q, want %q", request.request.SortOrder, request.wantSortOrder)
			}
		})
	}
}

func validUpdateAttendanceRequest() attendance.UpdateAttendanceRequest {
	return attendance.UpdateAttendanceRequest{}
}

func makeUpdateAttendanceRequest(modify func(*attendance.UpdateAttendanceRequest)) attendance.UpdateAttendanceRequest {
	req := validUpdateAttendanceRequest()
	if modify != nil {
		modify(&req)
	}
	return req
}

func TestUpdateAttendanceRequest(t *testing.T) {
	requests := []struct {
		name          string
		request       attendance.UpdateAttendanceRequest
		wantErrFields []string
	}{
		{
			name:          "empty request is valid (all fields optional)",
			request:       validUpdateAttendanceRequest(),
			wantErrFields: nil,
		},
		{
			name: "date format is not YYYY-MM-DD",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.Date = ptr("months")
			}),
			wantErrFields: []string{"date"},
		},
		{
			name: "date format is valid YYYY-MM-DD",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.Date = ptr("2026-01-01")
			}),
			wantErrFields: nil,
		},
		{
			name: "clock in time is YYYY-MM-DD HH:MM:SS",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockInTime = ptr("2026-05-12 12:15:12")
			}),
			wantErrFields: nil,
		},
		{
			name: "clock in time is HH:MM:SS",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockInTime = ptr("12:15:12")
			}),
			wantErrFields: nil,
		},
		{
			name: "clock in time format is not the allowed",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockInTime = ptr("2026-12-01")
			}),
			wantErrFields: []string{"clock_in_time"},
		},
		{
			name: "clock in time format is garbage",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockInTime = ptr("garbage")
			}),
			wantErrFields: []string{"clock_in_time"},
		},
		{
			name: "clock out time is YYYY-MM-DD HH:MM:SS",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockOutTime = ptr("2026-05-12 12:15:12")
			}),
			wantErrFields: nil,
		},
		{
			name: "clock out time is HH:MM:SS",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockOutTime = ptr("12:15:12")
			}),
			wantErrFields: nil,
		},
		{
			name: "clock out time format is not the allowed",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockOutTime = ptr("2026-12-01")
			}),
			wantErrFields: []string{"clock_out_time"},
		},
		{
			name: "clock out time format is garbage",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockOutTime = ptr("garbage")
			}),
			wantErrFields: []string{"clock_out_time"},
		},
		{
			name: "status value is not one of the allowed",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.Status = ptr("sigma")
			}),
			wantErrFields: []string{"status"},
		},
		{
			name: "status value is one of the allowed",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.Status = ptr("late")
			}),
			wantErrFields: nil,
		},
		{
			name: "status value is valid even with uppercase",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.Status = ptr("PRESENT")
			}),
			wantErrFields: nil,
		},
		{
			name: "status is nil skips validation",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.Status = nil
			}),
			wantErrFields: nil,
		},
		{
			name: "clock_in_latitude out of range",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockInLatitude = ptr(95.0)
			}),
			wantErrFields: []string{"clock_in_latitude"},
		},
		{
			name: "clock_in_latitude valid",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockInLatitude = ptr(70.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "clock_in_longitude out of range",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockInLongitude = ptr(185.0)
			}),
			wantErrFields: []string{"clock_in_longitude"},
		},
		{
			name: "clock_in_longitude valid",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockInLongitude = ptr(150.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "clock_out_latitude out of range",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockOutLatitude = ptr(-95.0)
			}),
			wantErrFields: []string{"clock_out_latitude"},
		},
		{
			name: "clock_out_latitude valid",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockOutLatitude = ptr(-70.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "clock_out_longitude out of range",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockOutLongitude = ptr(-185.0)
			}),
			wantErrFields: []string{"clock_out_longitude"},
		},
		{
			name: "clock_out_longitude valid",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.ClockOutLongitude = ptr(-150.0)
			}),
			wantErrFields: nil,
		},
		{
			name: "multiple validation errors accumulate",
			request: makeUpdateAttendanceRequest(func(uar *attendance.UpdateAttendanceRequest) {
				uar.Date = ptr("months")
				uar.Status = ptr("sigma")
				uar.ClockInLatitude = ptr(95.0)
				uar.ClockInLongitude = ptr(185.0)
				uar.ClockOutLatitude = ptr(-95.0)
				uar.ClockOutLongitude = ptr(-185.0)
			}),
			wantErrFields: []string{"date", "status", "clock_in_latitude", "clock_in_longitude", "clock_out_latitude", "clock_out_longitude"},
		},
	}

	for _, request := range requests {
		t.Run(request.name, func(t *testing.T) {
			err := request.request.Validate()
			assertValidationErrors(t, err, request.wantErrFields)
		})
	}
}

func validRejectAttendanceRequest() attendance.RejectAttendanceRequest {
	return attendance.RejectAttendanceRequest{
		ID:     "12345",
		Reason: "work location is too far than allowed",
	}
}

func makeRejectAttendanceRequest(modify func(*attendance.RejectAttendanceRequest)) attendance.RejectAttendanceRequest {
	req := validRejectAttendanceRequest()
	if modify != nil {
		modify(&req)
	}
	return req
}

func TestRejectAttendanceRequest(t *testing.T) {
	requests := []struct {
		name          string
		request       attendance.RejectAttendanceRequest
		wantErrFields []string
	}{
		{
			name:          "valid request",
			request:       validRejectAttendanceRequest(),
			wantErrFields: nil,
		},
		{
			name: "rejection reason is empty",
			request: makeRejectAttendanceRequest(
				func(rar *attendance.RejectAttendanceRequest) {
					rar.Reason = ""
				},
			),
			wantErrFields: []string{"reason"},
		},
		{
			name: "rejection reason is empty with trailing spaces",
			request: makeRejectAttendanceRequest(
				func(rar *attendance.RejectAttendanceRequest) {
					rar.Reason = "    "
				},
			),
			wantErrFields: []string{"reason"},
		},
		{
			name: "rejection reason value between trailing spaces",
			request: makeRejectAttendanceRequest(
				func(rar *attendance.RejectAttendanceRequest) {
					rar.Reason = "  reason  "
				},
			),
			wantErrFields: nil,
		},
	}

	for _, request := range requests {
		t.Run(request.name, func(t *testing.T) {
			err := request.request.Validate()
			assertValidationErrors(t, err, request.wantErrFields)
		})
	}
}
