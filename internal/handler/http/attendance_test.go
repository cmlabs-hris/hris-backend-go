package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	domainAttendance "github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	handlerAttendance "github.com/cmlabs-hris/hris-backend-go/internal/handler/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func testContext(t *testing.T, employee *string, company *string) context.Context {
	t.Helper()
	ja := jwtauth.New("HS256", []byte("test-secret"), nil)
	claims := make(map[string]interface{})
	if employee == nil {
		employee = new("emp-123")
	}
	if company == nil {
		company = new("company-123")
	}
	claims["employee_id"] = *employee
	claims["company_id"] = *company
	token, _, err := ja.Encode(claims)
	if err != nil {
		t.Fatalf("failed to encode claim: %v", err)
	}
	return jwtauth.NewContext(context.Background(), token, err)
}

type testAttendanceService struct {
	domainAttendance.AttendanceService
	getAttendanceFunc    func(ctx context.Context, id string) (attendance.AttendanceResponse, error)
	rejectAttendanceFunc func(ctx context.Context, req attendance.RejectAttendanceRequest) (attendance.AttendanceResponse, error)
}

func (tas *testAttendanceService) GetAttendance(ctx context.Context, id string) (attendance.AttendanceResponse, error) {
	return tas.getAttendanceFunc(ctx, id)
}

func (tas *testAttendanceService) RejectAttendance(ctx context.Context, req attendance.RejectAttendanceRequest) (attendance.AttendanceResponse, error) {
	return tas.rejectAttendanceFunc(ctx, req)
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		wantStatusCode int
		employee       *string
		company        *string
		result         attendance.AttendanceResponse
		want           func(t *testing.T, body *bytes.Buffer)
	}{
		{
			name:           "valid request",
			wantStatusCode: 200,
			result: attendance.AttendanceResponse{
				ID:         "1",
				EmployeeID: "emp-123",
				Date:       "2026-01-01",
				Status:     "present",
			},
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()
				type APIResponse struct {
					Success bool                          `json:"success"`
					Data    attendance.AttendanceResponse `json:"data"`
				}
				var response APIResponse
				if err := json.NewDecoder(body).Decode(&response); err != nil {
					t.Fatalf("failed to decode json: %v", err)
				}
				if response.Data.ID != "1" {
					t.Errorf("want ID=1 but instead got %s", response.Data.ID)
				}
				if response.Data.EmployeeID != "emp-123" {
					t.Errorf("want EmployeeID=emp-123 but instead got %s", response.Data.EmployeeID)
				}
				if response.Data.Date != "2026-01-01" {
					t.Errorf("want Date=2026-01-01 but instead got %s", response.Data.Date)
				}
				if response.Data.Status != "present" {
					t.Errorf("want Status=present but instead got %s", response.Data.Status)
				}
			},
		},
		{
			name:           "generic service error returns 500",
			wantStatusCode: 500,
			serviceErr:     errors.New("failed to extract claims from context"),
			result:         attendance.AttendanceResponse{},
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()
				if !strings.Contains(body.String(), "An unexpected error occurred") {
					t.Errorf("want message field to contain 'An unexpected error occurred' but instead got %s", body.String())
				}

			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attSrv := testAttendanceService{
				getAttendanceFunc: func(ctx context.Context, id string) (domainAttendance.AttendanceResponse, error) {
					return test.result, test.serviceErr
				},
			}
			req := httptest.NewRequest(http.MethodGet, "/attendances/1", nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			handler := handlerAttendance.NewAttendanceHandler(&attSrv)
			r.Get("/attendances/{id}", handler.Get)
			r.ServeHTTP(rec, req)

			if rec.Code != test.wantStatusCode {
				t.Errorf("want %d status code, but instead got %d", test.wantStatusCode, rec.Code)
			}

			test.want(t, rec.Body)
		})
	}
}

func TestRejectAttendance(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		wantStatusCode int
		request        attendance.RejectAttendanceRequest
		want           func(t *testing.T, body *bytes.Buffer)
	}{
		{
			name:           "valid request",
			wantStatusCode: http.StatusOK,
			request: attendance.RejectAttendanceRequest{
				ID:     "att-123",
				Reason: "busy",
			},
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()
				type APIResponse struct {
					Success bool                          `json:"success"`
					Data    attendance.AttendanceResponse `json:"data"`
				}
				var response APIResponse
				if err := json.Unmarshal(body.Bytes(), &response); err != nil {
					t.Fatalf("failed to parse json: %v", err)
				}
				if response.Data.ID != "att-123" {
					t.Errorf("want ID=att-123 but instead got %s", response.Data.ID)
				}
				fmt.Println(response.Data.RejectionReason)
				if *response.Data.RejectionReason != "busy" {
					t.Errorf("want RejectionReason=busy but instead got %s", *response.Data.RejectionReason)
				}
			},
		},
		{
			name:           "dto validation error",
			wantStatusCode: 422,
			request: domainAttendance.RejectAttendanceRequest{
				ID:     "att-123",
				Reason: "",
			},

			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()
				if !strings.Contains(body.String(), `"code":"VALIDATION_ERROR"`) {
					t.Errorf(`want message field to contain "code":"VALIDATION_ERROR" but instead got %s`, body.String())
				}
			},
		},
		{
			name:           "attendance not found",
			wantStatusCode: 404,
			request: domainAttendance.RejectAttendanceRequest{
				ID:     "att-123",
				Reason: "something",
			},
			serviceErr: attendance.ErrAttendanceNotFound,
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()
				if !strings.Contains(body.String(), `"code":"NOT_FOUND"`) {
					t.Errorf(`want message field to contain "code":"NOT_FOUND" but instead got %s`, body.String())
				}
			},
		},
		{
			name:           "attendance has already been processed",
			wantStatusCode: 409,
			request: domainAttendance.RejectAttendanceRequest{
				ID:     "att-123",
				Reason: "something",
			},
			serviceErr: attendance.ErrAttendanceAlreadyProcessed,
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()
				if !strings.Contains(body.String(), `"code":"CONFLICT"`) {
					t.Errorf(`want message field to contain "code":"CONFLICT" but instead got %s`, body.String())
				}
			},
		},
		{
			name:           "error repository db down",
			wantStatusCode: 500,
			request: domainAttendance.RejectAttendanceRequest{
				ID:     "att-123",
				Reason: "something",
			},
			serviceErr: errors.New("failed to get attendance: db down"),
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()

				if !strings.Contains(body.String(), `"code":"INTERNAL_SERVER_ERROR"`) {
					t.Errorf(`want message field to contain "code":"INTERNAL_SERVER_ERROR" but instead got %s`, body.String())
				}
			},
		},
		{
			name:           "attendance has been approved",
			wantStatusCode: 409,
			request: domainAttendance.RejectAttendanceRequest{
				ID:     "att-123",
				Reason: "something",
			},
			serviceErr: attendance.ErrCannotRejectApproved,
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()
				if !strings.Contains(body.String(), `"code":"CONFLICT"`) {
					t.Errorf(`want message field to contain "code":"CONFLICT" but instead got %s`, body.String())
				}
			},
		},
		{
			name:           "repository attendance update error db down",
			wantStatusCode: 500,
			request: domainAttendance.RejectAttendanceRequest{
				ID:     "att-123",
				Reason: "something",
			},
			serviceErr: errors.New("failed to reject attendance: db down"),
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()

				if !strings.Contains(body.String(), `"code":"INTERNAL_SERVER_ERROR"`) {
					t.Errorf(`want message field to contain "code":"INTERNAL_SERVER_ERROR" but instead got %s`, body.String())
				}
			},
		},
		{
			name:           "repository attendance getbyid error db down",
			wantStatusCode: 500,
			request: domainAttendance.RejectAttendanceRequest{
				ID:     "att-123",
				Reason: "something",
			},
			serviceErr: errors.New("failed to get updated attendance: db down"),
			want: func(t *testing.T, body *bytes.Buffer) {
				t.Helper()

				if !strings.Contains(body.String(), `"code":"INTERNAL_SERVER_ERROR"`) {
					t.Errorf(`want message field to contain "code":"INTERNAL_SERVER_ERROR" but instead got %s`, body.String())
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(test.request)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}
			fmt.Println(string(jsonBytes))
			rec := httptest.NewRecorder()
			srv := testAttendanceService{
				rejectAttendanceFunc: func(ctx context.Context, req domainAttendance.RejectAttendanceRequest) (domainAttendance.AttendanceResponse, error) {
					return domainAttendance.AttendanceResponse{
						ID:              test.request.ID,
						RejectionReason: &test.request.Reason,
					}, test.serviceErr
				},
			}
			handler := handlerAttendance.NewAttendanceHandler(&srv)

			r := chi.NewRouter()
			r.Post("/{id}/reject", handler.Reject)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s/reject", test.request.ID), bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(rec, req)

			if rec.Code != test.wantStatusCode {
				t.Errorf("want status code %d, but instead got %d", test.wantStatusCode, rec.Code)
			}
			fmt.Println("hellaur")
			fmt.Println(rec.Code)
			test.want(t, rec.Body)
		})
	}
}
