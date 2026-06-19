const fs = require('fs');

const collection = {
    info: {
        _postman_id: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        name: "HRIS cmlabs API",
        schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
        description: "Postman collection for the HRIS cmlabs backend API.\n\nVariables:\n- baseUrl: API base URL (default: http://localhost:8080)\n- accessToken: JWT access token (auto-set on login/register)\n- refreshToken: JWT refresh token (auto-set on login/register)\n- companyId, employeeId, branchId, gradeId, positionId, scheduleId, leaveTypeId, leaveRequestId, attendanceId, payrollComponentId, notificationId, invoiceId, planId, invitationToken"
    },
    variable: [
        {key: "baseUrl", value: "http://localhost:8080"},
        {key: "accessToken", value: ""},
        {key: "refreshToken", value: ""},
        {key: "companyId", value: ""},
        {key: "employeeId", value: ""},
        {key: "branchId", value: ""},
        {key: "gradeId", value: ""},
        {key: "positionId", value: ""},
        {key: "scheduleId", value: ""},
        {key: "scheduleTimeId", value: ""},
        {key: "scheduleLocationId", value: ""},
        {key: "scheduleAssignId", value: ""},
        {key: "leaveTypeId", value: ""},
        {key: "leaveQuotaId", value: ""},
        {key: "leaveRequestId", value: ""},
        {key: "attendanceId", value: ""},
        {key: "payrollComponentId", value: ""},
        {key: "employeeComponentId", value: ""},
        {key: "payrollRecordId", value: ""},
        {key: "notificationId", value: ""},
        {key: "invoiceId", value: ""},
        {key: "planId", value: ""},
        {key: "invitationToken", value: ""},
        {key: "sseToken", value: ""}
    ],
    auth: {
        type: "bearer",
        bearer: [{key: "token", value: "{{accessToken}}", type: "string"}]
    },
    item: [
        {
            name: "Auth",
            item: [
                {
                    name: "Register",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}],
                        url: {raw: "{{baseUrl}}/api/v1/auth/register", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "register"]},
                        body: {mode: "raw", raw: "{\n  \"email\": \"owner@acme.com\",\n  \"password\": \"Password123!\",\n  \"confirm_password\": \"Password123!\"\n}"}
                    },
                    event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data) {", "  if (res.data.access_token) pm.collectionVariables.set('accessToken', res.data.access_token);", "  if (res.data.refresh_token) pm.collectionVariables.set('refreshToken', res.data.refresh_token);", "}"]}}]
                },
                {
                    name: "Login",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}],
                        url: {raw: "{{baseUrl}}/api/v1/auth/login", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "login"]},
                        body: {mode: "raw", raw: "{\n  \"email\": \"owner@acme.com\",\n  \"password\": \"Password123!\"\n}"}
                    },
                    event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data) {", "  if (res.data.access_token) pm.collectionVariables.set('accessToken', res.data.access_token);", "  if (res.data.refresh_token) pm.collectionVariables.set('refreshToken', res.data.refresh_token);", "}"]}}]
                },
                {
                    name: "Login with Employee Code",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}],
                        url: {raw: "{{baseUrl}}/api/v1/auth/login/employee-code", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "login", "employee-code"]},
                        body: {mode: "raw", raw: "{\n  \"company_username\": \"acme-corp\",\n  \"employee_code\": \"EMP-001\",\n  \"password\": \"Password123!\"\n}"}
                    },
                    event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data) {", "  if (res.data.access_token) pm.collectionVariables.set('accessToken', res.data.access_token);", "  if (res.data.refresh_token) pm.collectionVariables.set('refreshToken', res.data.refresh_token);", "}"]}}]
                },
                {
                    name: "Login OAuth Google (Browser)",
                    request: {method: "GET", header: [], url: {raw: "{{baseUrl}}/api/v1/auth/login/oauth/google", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "login", "oauth", "google"]}}
                },
                {
                    name: "Refresh Token",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}],
                        url: {raw: "{{baseUrl}}/api/v1/auth/refresh", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "refresh"]},
                        body: {mode: "raw", raw: "{\n  \"refresh_token\": \"{{refreshToken}}\"\n}"}
                    },
                    event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data) {", "  if (res.data.access_token) pm.collectionVariables.set('accessToken', res.data.access_token);", "}"]}}]
                },
                {
                    name: "Logout",
                    request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/auth/logout", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "logout"]}},
                    event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success) {", "  pm.collectionVariables.set('accessToken', '');", "  pm.collectionVariables.set('refreshToken', '');", "}"]}}]
                },
                {
                    name: "Forgot Password",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}],
                        url: {raw: "{{baseUrl}}/api/v1/auth/forgot-password", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "forgot-password"]},
                        body: {mode: "raw", raw: "{\n  \"email\": \"owner@acme.com\"\n}"}
                    }
                },
                {
                    name: "Reset Password",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}],
                        url: {raw: "{{baseUrl}}/api/v1/auth/reset-password", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "reset-password"]},
                        body: {mode: "raw", raw: "{\n  \"token\": \"reset-token-from-email\",\n  \"new_password\": \"NewPassword123!\",\n  \"confirm_new_password\": \"NewPassword123!\"\n}"}
                    }
                },
                {
                    name: "Verify Email",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}],
                        url: {raw: "{{baseUrl}}/api/v1/auth/verify-email", host: ["{{baseUrl}}"], path: ["api", "v1", "auth", "verify-email"]},
                        body: {mode: "raw", raw: "{\n  \"token\": \"verification-token-from-email\"\n}"}
                    }
                }
            ]
        },
        {
            name: "Company",
            item: [
                {
                    name: "Create Company",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}],
                        url: {raw: "{{baseUrl}}/api/v1/company", host: ["{{baseUrl}}"], path: ["api", "v1", "company"]},
                        body: {mode: "raw", raw: "{\n  \"name\": \"PT Acme Corp\",\n  \"npwp\": \"01.234.567.8-901.000\",\n  \"address\": \"Jl. Sudirman No. 1, Jakarta\",\n  \"phone\": \"+62211234567\",\n  \"email\": \"info@acme.com\",\n  \"website\": \"https://acme.com\"\n}"}
                    }
                },
                {
                    name: "Get My Company",
                    request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/company/my", host: ["{{baseUrl}}"], path: ["api", "v1", "company", "my"]}}
                },
                {
                    name: "Update My Company",
                    request: {
                        method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}],
                        url: {raw: "{{baseUrl}}/api/v1/company/my", host: ["{{baseUrl}}"], path: ["api", "v1", "company", "my"]},
                        body: {mode: "raw", raw: "{\n  \"name\": \"PT Acme Corp Updated\",\n  \"address\": \"Jl. Thamrin No. 10, Jakarta\"\n}"}
                    }
                },
                {
                    name: "Delete My Company",
                    request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/company/my", host: ["{{baseUrl}}"], path: ["api", "v1", "company", "my"]}}
                },
                {
                    name: "Upload Company Logo",
                    request: {
                        method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}],
                        url: {raw: "{{baseUrl}}/api/v1/company/my/logo", host: ["{{baseUrl}}"], path: ["api", "v1", "company", "my", "logo"]},
                        body: {mode: "formdata", formdata: [{key: "logo", type: "file", src: ""}]}
                    }
                }
            ]
        },
        {
            name: "Master - Branches",
            item: [
                {name: "List Branches", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/branches", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "branches"]}}},
                {name: "Get Branch", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/branches/{{branchId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "branches", "{{branchId}}"]}}},
                {
                    name: "Create Branch",
                    request: {
                        method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}],
                        url: {raw: "{{baseUrl}}/api/v1/master/branches", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "branches"]},
                        body: {mode: "raw", raw: "{\n  \"name\": \"HQ Jakarta\",\n  \"address\": \"Jl. Sudirman No. 1\"\n}"}
                    },
                    event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) {", "  pm.collectionVariables.set('branchId', res.data.id);", "}"]}}]
                },
                {
                    name: "Update Branch",
                    request: {
                        method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}],
                        url: {raw: "{{baseUrl}}/api/v1/master/branches/{{branchId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "branches", "{{branchId}}"]},
                        body: {mode: "raw", raw: "{\n  \"name\": \"HQ Jakarta Updated\",\n  \"timezone\": \"Asia/Jakarta\"\n}"}
                    }
                },
                {name: "Delete Branch", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/branches/{{branchId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "branches", "{{branchId}}"]}}}
            ]
        },
        {
            name: "Master - Grades",
            item: [
                {name: "List Grades", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/grades", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "grades"]}}},
                {name: "Get Grade", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/grades/{{gradeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "grades", "{{gradeId}}"]}}},
                {name: "Create Grade", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/grades", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "grades"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Senior\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('gradeId', res.data.id);"]}}]},
                {name: "Update Grade", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/grades/{{gradeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "grades", "{{gradeId}}"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Senior Updated\"\n}"}}},
                {name: "Delete Grade", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/grades/{{gradeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "grades", "{{gradeId}}"]}}}
            ]
        },
        {
            name: "Master - Positions",
            item: [
                {name: "List Positions", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/positions", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "positions"]}}},
                {name: "Get Position", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/positions/{{positionId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "positions", "{{positionId}}"]}}},
                {name: "Create Position", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/positions", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "positions"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Software Engineer\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('positionId', res.data.id);"]}}]},
                {name: "Update Position", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/positions/{{positionId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "positions", "{{positionId}}"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Senior Software Engineer\"\n}"}}},
                {name: "Delete Position", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/master/positions/{{positionId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "master", "positions", "{{positionId}}"]}}}
            ]
        },
        {
            name: "Employee",
            item: [
                {name: "List Employees", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees?page=1&limit=10", host: ["{{baseUrl}}"], path: ["api", "v1", "employees"], query: [{key: "page", value: "1"}, {key: "limit", value: "10"}]}}},
                {name: "Search Employees", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees/search?q=john&limit=10", host: ["{{baseUrl}}"], path: ["api", "v1", "employees", "search"], query: [{key: "q", value: "john"}, {key: "limit", value: "10"}]}}},
                {
                    name: "Create Employee",
                    request: {
                        method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}],
                        url: {raw: "{{baseUrl}}/api/v1/employees", host: ["{{baseUrl}}"], path: ["api", "v1", "employees"]},
                        body: {
                            mode: "formdata",
                            formdata: [
                                {key: "first_name", value: "John", type: "text"},
                                {key: "last_name", value: "Doe", type: "text"},
                                {key: "email", value: "john.doe@acme.com", type: "text"},
                                {key: "phone", value: "+6281234567890", type: "text"},
                                {key: "gender", value: "male", type: "text"},
                                {key: "birth_date", value: "1995-01-15", type: "text"},
                                {key: "position_id", value: "{{positionId}}", type: "text"},
                                {key: "branch_id", value: "{{branchId}}", type: "text"},
                                {key: "grade_id", value: "{{gradeId}}", type: "text"},
                                {key: "employment_type", value: "permanent", type: "text"},
                                {key: "join_date", value: "2026-01-01", type: "text"},
                                {key: "base_salary", value: "10000000", type: "text"},
                                {key: "role", value: "employee", type: "text"},
                                {key: "avatar", type: "file", src: ""}
                            ]
                        }
                    },
                    event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('employeeId', res.data.id);"]}}]
                },
                {name: "Get Employee", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees/{{employeeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "employees", "{{employeeId}}"]}}},
                {name: "Update Employee", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees/{{employeeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "employees", "{{employeeId}}"]}, body: {mode: "raw", raw: "{\n  \"first_name\": \"John\",\n  \"last_name\": \"Doe Updated\",\n  \"phone\": \"+6281234567891\"\n}"}}},
                {name: "Upload Employee Avatar", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees/{{employeeId}}/avatar", host: ["{{baseUrl}}"], path: ["api", "v1", "employees", "{{employeeId}}", "avatar"]}, body: {mode: "formdata", formdata: [{key: "avatar", type: "file", src: ""}]}}},
                {name: "Inactivate Employee", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees/{{employeeId}}/inactivate", host: ["{{baseUrl}}"], path: ["api", "v1", "employees", "{{employeeId}}", "inactivate"]}}},
                {name: "Resend Invitation", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees/{{employeeId}}/invitation/resend", host: ["{{baseUrl}}"], path: ["api", "v1", "employees", "{{employeeId}}", "invitation", "resend"]}}},
                {name: "Revoke Invitation", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees/{{employeeId}}/invitation/revoke", host: ["{{baseUrl}}"], path: ["api", "v1", "employees", "{{employeeId}}", "invitation", "revoke"]}}},
                {name: "Delete Employee", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employees/{{employeeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "employees", "{{employeeId}}"]}}}
            ]
        },
        {
            name: "Invitation",
            item: [
                {name: "View Invitation (Public)", request: {method: "GET", header: [], url: {raw: "{{baseUrl}}/api/v1/invitations/view/{{invitationToken}}", host: ["{{baseUrl}}"], path: ["api", "v1", "invitations", "view", "{{invitationToken}}"]}}},
                {name: "List My Invitations", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/invitations/my", host: ["{{baseUrl}}"], path: ["api", "v1", "invitations", "my"]}}},
                {name: "Accept Invitation", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/invitations/{{invitationToken}}/accept", host: ["{{baseUrl}}"], path: ["api", "v1", "invitations", "{{invitationToken}}", "accept"]}}}
            ]
        },
        {
            name: "Leave - Types",
            item: [
                {name: "List Leave Types", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/types", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "types"]}}},
                {name: "Create Leave Type", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/types", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "types"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Annual Leave\",\n  \"code\": \"AL\",\n  \"default_quota\": 12,\n  \"is_paid\": true,\n  \"is_carry_forward\": false,\n  \"require_attachment\": false,\n  \"min_days_notice\": 3,\n  \"applicable_gender\": \"all\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('leaveTypeId', res.data.id);"]}}]},
                {name: "Update Leave Type", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/types/{{leaveTypeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "types", "{{leaveTypeId}}"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Annual Leave Updated\",\n  \"default_quota\": 14\n}"}}},
                {name: "Delete Leave Type", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/types/{{leaveTypeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "types", "{{leaveTypeId}}"]}}}
            ]
        },
        {
            name: "Leave - Quota",
            item: [
                {name: "List Quota (Manager)", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/quota?page=1&limit=10", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "quota"], query: [{key: "page", value: "1"}, {key: "limit", value: "10"}]}}},
                {name: "Get My Quota", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/quota/my", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "quota", "my"]}}},
                {name: "Get Quota by ID", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/quota/{{leaveQuotaId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "quota", "{{leaveQuotaId}}"]}}},
                {name: "Adjust Quota", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/quota/adjust", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "quota", "adjust"]}, body: {mode: "raw", raw: "{\n  \"employee_id\": \"{{employeeId}}\",\n  \"leave_type_id\": \"{{leaveTypeId}}\",\n  \"adjustment\": 2,\n  \"reason\": \"Extra leave for good performance\"\n}"}}}
            ]
        },
        {
            name: "Leave - Requests",
            item: [
                {name: "List Requests (Manager)", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/requests?page=1&limit=10", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "requests"], query: [{key: "page", value: "1"}, {key: "limit", value: "10"}]}}},
                {name: "Get My Requests", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/requests/my", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "requests", "my"]}}},
                {name: "Create Leave Request", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/requests", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "requests"]}, body: {mode: "raw", raw: "{\n  \"leave_type_id\": \"{{leaveTypeId}}\",\n  \"start_date\": \"2026-07-01\",\n  \"end_date\": \"2026-07-03\",\n  \"reason\": \"Family vacation\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('leaveRequestId', res.data.id);"]}}]},
                {name: "Get Request by ID", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/requests/{{leaveRequestId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "requests", "{{leaveRequestId}}"]}}},
                {name: "Approve Leave Request", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/requests/{{leaveRequestId}}/approve", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "requests", "{{leaveRequestId}}", "approve"]}}},
                {name: "Reject Leave Request", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/leave/requests/{{leaveRequestId}}/reject", host: ["{{baseUrl}}"], path: ["api", "v1", "leave", "requests", "{{leaveRequestId}}", "reject"]}, body: {mode: "raw", raw: "{\n  \"reason\": \"Team is understaffed during that period\"\n}"}}}
            ]
        },
        {
            name: "Schedule",
            item: [
                {name: "List Work Schedules", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule"]}}},
                {name: "Create Work Schedule", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Regular 9-5\",\n  \"type\": \"fixed\",\n  \"effective_date\": \"2026-01-01\",\n  \"is_default\": true\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('scheduleId', res.data.id);"]}}]},
                {name: "Get Work Schedule", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/{{scheduleId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "{{scheduleId}}"]}}},
                {name: "Update Work Schedule", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/{{scheduleId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "{{scheduleId}}"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Regular 9-5 Updated\"\n}"}}},
                {name: "Delete Work Schedule", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/{{scheduleId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "{{scheduleId}}"]}}},
                {name: "Get Employee Schedule Timeline", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/employee/{{employeeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "employee", "{{employeeId}}"]}}},
                {name: "Assign Schedule to Employee", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/{{scheduleId}}/employee/{{employeeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "{{scheduleId}}", "employee", "{{employeeId}}"]}, body: {mode: "raw", raw: "{\n  \"effective_date\": \"2026-01-01\"\n}"}}},
                {name: "Update Employee Schedule Assignment", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/{{scheduleAssignId}}/employee/{{employeeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "{{scheduleAssignId}}", "employee", "{{employeeId}}"]}, body: {mode: "raw", raw: "{\n  \"effective_date\": \"2026-02-01\"\n}"}}},
                {name: "Delete Employee Schedule Assignment", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/{{scheduleAssignId}}/employee/{{employeeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "{{scheduleAssignId}}", "employee", "{{employeeId}}"]}}},
                {name: "Create Schedule Time", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/times", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "times"]}, body: {mode: "raw", raw: "{\n  \"work_schedule_id\": \"{{scheduleId}}\",\n  \"day\": \"monday\",\n  \"start_time\": \"09:00\",\n  \"end_time\": \"17:00\",\n  \"is_work_day\": true,\n  \"break_start\": \"12:00\",\n  \"break_end\": \"13:00\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('scheduleTimeId', res.data.id);"]}}]},
                {name: "Get Schedule Time", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/times/{{scheduleTimeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "times", "{{scheduleTimeId}}"]}}},
                {name: "Update Schedule Time", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/times/{{scheduleTimeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "times", "{{scheduleTimeId}}"]}, body: {mode: "raw", raw: "{\n  \"start_time\": \"08:00\",\n  \"end_time\": \"16:00\"\n}"}}},
                {name: "Delete Schedule Time", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/times/{{scheduleTimeId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "times", "{{scheduleTimeId}}"]}}},
                {name: "Create Schedule Location", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/locations", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "locations"]}, body: {mode: "raw", raw: "{\n  \"work_schedule_id\": \"{{scheduleId}}\",\n  \"name\": \"Office Jakarta\",\n  \"latitude\": -6.2088,\n  \"longitude\": 106.8456,\n  \"radius_meters\": 100,\n  \"address\": \"Jl. Sudirman No. 1\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('scheduleLocationId', res.data.id);"]}}]},
                {name: "Get Schedule Location", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/locations/{{scheduleLocationId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "locations", "{{scheduleLocationId}}"]}}},
                {name: "Update Schedule Location", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/locations/{{scheduleLocationId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "locations", "{{scheduleLocationId}}"]}, body: {mode: "raw", raw: "{\n  \"radius_meters\": 200\n}"}}},
                {name: "Delete Schedule Location", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/schedule/locations/{{scheduleLocationId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "schedule", "locations", "{{scheduleLocationId}}"]}}}
            ]
        },
        {
            name: "Employee Schedules",
            item: [
                {name: "List Assignments", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employee-schedules?page=1&limit=10", host: ["{{baseUrl}}"], path: ["api", "v1", "employee-schedules"], query: [{key: "page", value: "1"}, {key: "limit", value: "10"}]}}},
                {name: "Get Active Schedule", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employee-schedules/active", host: ["{{baseUrl}}"], path: ["api", "v1", "employee-schedules", "active"]}}},
                {name: "Get Assignment by ID", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employee-schedules/{{scheduleAssignId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "employee-schedules", "{{scheduleAssignId}}"]}}},
                {name: "Create Assignment", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employee-schedules", host: ["{{baseUrl}}"], path: ["api", "v1", "employee-schedules"]}, body: {mode: "raw", raw: "{\n  \"employee_id\": \"{{employeeId}}\",\n  \"work_schedule_id\": \"{{scheduleId}}\",\n  \"effective_date\": \"2026-01-01\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('scheduleAssignId', res.data.id);"]}}]},
                {name: "Update Assignment", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employee-schedules/{{scheduleAssignId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "employee-schedules", "{{scheduleAssignId}}"]}, body: {mode: "raw", raw: "{\n  \"effective_date\": \"2026-02-01\"\n}"}}},
                {name: "Delete Assignment", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/employee-schedules/{{scheduleAssignId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "employee-schedules", "{{scheduleAssignId}}"]}}}
            ]
        },
        {
            name: "Attendance",
            item: [
                {name: "Get My Attendance", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/my?page=1&limit=10", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "my"], query: [{key: "page", value: "1"}, {key: "limit", value: "10"}]}}},
                {name: "Get Attendance Status", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/status", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "status"]}}},
                {name: "Clock In", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/clock-in", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "clock-in"]}, body: {mode: "formdata", formdata: [{key: "latitude", value: "-6.2088", type: "text"}, {key: "longitude", value: "106.8456", type: "text"}, {key: "photo", type: "file", src: ""}]}}},
                {name: "Clock Out", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/clock-out", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "clock-out"]}, body: {mode: "formdata", formdata: [{key: "latitude", value: "-6.2088", type: "text"}, {key: "longitude", value: "106.8456", type: "text"}, {key: "photo", type: "file", src: ""}]}}},
                {name: "List All Attendance (Manager)", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance?page=1&limit=10", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance"], query: [{key: "page", value: "1"}, {key: "limit", value: "10"}]}}},
                {name: "Get Attendance by ID", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/{{attendanceId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "{{attendanceId}}"]}}},
                {name: "Update Attendance", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/{{attendanceId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "{{attendanceId}}"]}, body: {mode: "raw", raw: "{\n  \"clock_in\": \"2026-06-19T09:00:00Z\",\n  \"clock_out\": \"2026-06-19T17:00:00Z\",\n  \"notes\": \"Corrected time\"\n}"}}},
                {name: "Approve Attendance", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/{{attendanceId}}/approve", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "{{attendanceId}}", "approve"]}}},
                {name: "Reject Attendance", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/{{attendanceId}}/reject", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "{{attendanceId}}", "reject"]}, body: {mode: "raw", raw: "{\n  \"reason\": \"Invalid clock-in location\"\n}"}}},
                {name: "Delete Attendance", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/attendance/{{attendanceId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "attendance", "{{attendanceId}}"]}}}
            ]
        },
        {
            name: "Payroll",
            item: [
                {name: "Get Settings", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/settings", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "settings"]}}},
                {name: "Update Settings", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/settings", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "settings"]}, body: {mode: "raw", raw: "{\n  \"late_deduction_enabled\": true,\n  \"late_deduction_per_minute\": \"500.00\",\n  \"overtime_enabled\": true,\n  \"overtime_pay_per_minute\": \"750.00\"\n}"}}},
                {name: "List Components", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/components", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "components"]}}},
                {name: "Create Component", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/components", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "components"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Transport Allowance\",\n  \"type\": \"allowance\",\n  \"description\": \"Monthly transport allowance\",\n  \"is_taxable\": false\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('payrollComponentId', res.data.id);"]}}]},
                {name: "Get Component", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/components/{{payrollComponentId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "components", "{{payrollComponentId}}"]}}},
                {name: "Update Component", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/components/{{payrollComponentId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "components", "{{payrollComponentId}}"]}, body: {mode: "raw", raw: "{\n  \"name\": \"Transport Allowance Updated\"\n}"}}},
                {name: "Delete Component", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/components/{{payrollComponentId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "components", "{{payrollComponentId}}"]}}},
                {name: "Get Employee Components", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/employees/{{employeeId}}/components", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "employees", "{{employeeId}}", "components"]}}},
                {name: "Assign Component to Employee", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/employees/{{employeeId}}/components", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "employees", "{{employeeId}}", "components"]}, body: {mode: "raw", raw: "{\n  \"payroll_component_id\": \"{{payrollComponentId}}\",\n  \"amount\": \"500000.00\",\n  \"effective_date\": \"2026-01-01\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.id) pm.collectionVariables.set('employeeComponentId', res.data.id);"]}}]},
                {name: "Update Employee Component", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/employee-components/{{employeeComponentId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "employee-components", "{{employeeComponentId}}"]}, body: {mode: "raw", raw: "{\n  \"amount\": \"600000.00\"\n}"}}},
                {name: "Remove Employee Component", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/employee-components/{{employeeComponentId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "employee-components", "{{employeeComponentId}}"]}}},
                {name: "Generate Payroll", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/generate", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "generate"]}, body: {mode: "raw", raw: "{\n  \"period_month\": 6,\n  \"period_year\": 2026\n}"}}},
                {name: "List Payroll Records", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/records?page=1&limit=10&period_month=6&period_year=2026", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "records"], query: [{key: "page", value: "1"}, {key: "limit", value: "10"}, {key: "period_month", value: "6"}, {key: "period_year", value: "2026"}]}}},
                {name: "Get Payroll Record", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/records/{{payrollRecordId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "records", "{{payrollRecordId}}"]}}},
                {name: "Update Payroll Record", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/records/{{payrollRecordId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "records", "{{payrollRecordId}}"]}, body: {mode: "raw", raw: "{\n  \"notes\": \"Adjusted overtime\"\n}"}}},
                {name: "Delete Payroll Record", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/records/{{payrollRecordId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "records", "{{payrollRecordId}}"]}}},
                {name: "Finalize Payroll", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/finalize", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "finalize"]}, body: {mode: "raw", raw: "{\n  \"record_ids\": [\"{{payrollRecordId}}\"]\n}"}}},
                {name: "Get Payroll Summary", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/payroll/summary?period_month=6&period_year=2026", host: ["{{baseUrl}}"], path: ["api", "v1", "payroll", "summary"], query: [{key: "period_month", value: "6"}, {key: "period_year", value: "2026"}]}}}
            ]
        },
        {
            name: "Dashboard - Admin",
            item: [
                {name: "Get Admin Dashboard", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/admin", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "admin"]}}},
                {name: "Employee Current Number", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/admin/employee-current-number?month=2026-06", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "admin", "employee-current-number"], query: [{key: "month", value: "2026-06"}]}}},
                {name: "Employee Status Stats", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/admin/employee-status-stats?month=2026-06", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "admin", "employee-status-stats"], query: [{key: "month", value: "2026-06"}]}}},
                {name: "Monthly Attendance", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/admin/monthly-attendance?month=2026-06", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "admin", "monthly-attendance"], query: [{key: "month", value: "2026-06"}]}}},
                {name: "Daily Attendance Stats", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/admin/daily-attendance-stats?date=2026-06-19", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "admin", "daily-attendance-stats"], query: [{key: "date", value: "2026-06-19"}]}}}
            ]
        },
        {
            name: "Dashboard - Employee",
            item: [
                {name: "Get Employee Dashboard", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/employee", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "employee"]}}},
                {name: "Work Stats", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/employee/work-stats?start_date=2026-06-01&end_date=2026-06-30", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "employee", "work-stats"], query: [{key: "start_date", value: "2026-06-01"}, {key: "end_date", value: "2026-06-30"}]}}},
                {name: "Attendance Summary", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/employee/attendance-summary?month=2026-06", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "employee", "attendance-summary"], query: [{key: "month", value: "2026-06"}]}}},
                {name: "Leave Summary", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/employee/leave-summary?year=2026", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "employee", "leave-summary"], query: [{key: "year", value: "2026"}]}}},
                {name: "Work Hours Chart", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/dashboard/employee/work-hours-chart?month=2026-06&week=3", host: ["{{baseUrl}}"], path: ["api", "v1", "dashboard", "employee", "work-hours-chart"], query: [{key: "month", value: "2026-06"}, {key: "week", value: "3"}]}}}
            ]
        },
        {
            name: "Notification",
            item: [
                {name: "List Notifications", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/notifications?page=1&page_size=10", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications"], query: [{key: "page", value: "1"}, {key: "page_size", value: "10"}]}}},
                {name: "Get Unread Count", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/notifications/unread-count", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications", "unread-count"]}}},
                {name: "Mark As Read", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/notifications/mark-read", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications", "mark-read"]}, body: {mode: "raw", raw: "{\n  \"notification_ids\": [\"{{notificationId}}\"]\n}"}}},
                {name: "Mark All As Read", request: {method: "POST", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/notifications/mark-all-read", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications", "mark-all-read"]}}},
                {name: "Delete Notification", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/notifications/{{notificationId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications", "{{notificationId}}"]}}},
                {name: "Get Preferences", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/notifications/preferences", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications", "preferences"]}}},
                {name: "Update Preference", request: {method: "PUT", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/notifications/preferences", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications", "preferences"]}, body: {mode: "raw", raw: "{\n  \"notification_type\": \"leave_approved\",\n  \"email_enabled\": true,\n  \"push_enabled\": true\n}"}}},
                {name: "Get SSE Token", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/notifications/token", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications", "token"]}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.token) pm.collectionVariables.set('sseToken', res.data.token);"]}}]},
                {name: "SSE Stream", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}, {key: "Accept", value: "text/event-stream"}], url: {raw: "{{baseUrl}}/api/v1/notifications/stream?token={{sseToken}}", host: ["{{baseUrl}}"], path: ["api", "v1", "notifications", "stream"], query: [{key: "token", value: "{{sseToken}}"}]}}}
            ]
        },
        {
            name: "Report",
            item: [
                {name: "Monthly Attendance Report", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/reports/attendance?month=6&year=2026", host: ["{{baseUrl}}"], path: ["api", "v1", "reports", "attendance"], query: [{key: "month", value: "6"}, {key: "year", value: "2026"}]}}},
                {name: "Payroll Summary Report", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/reports/payroll?month=6&year=2026", host: ["{{baseUrl}}"], path: ["api", "v1", "reports", "payroll"], query: [{key: "month", value: "6"}, {key: "year", value: "2026"}]}}},
                {name: "Leave Balance Report", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/reports/leave-balance?year=2026", host: ["{{baseUrl}}"], path: ["api", "v1", "reports", "leave-balance"], query: [{key: "year", value: "2026"}]}}},
                {name: "New Hire Report", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/reports/new-hires?start_date=2026-01-01&end_date=2026-06-30", host: ["{{baseUrl}}"], path: ["api", "v1", "reports", "new-hires"], query: [{key: "start_date", value: "2026-01-01"}, {key: "end_date", value: "2026-06-30"}]}}}
            ]
        },
        {
            name: "Subscription",
            item: [
                {name: "List Plans (Public)", request: {method: "GET", header: [], url: {raw: "{{baseUrl}}/api/v1/plans", host: ["{{baseUrl}}"], path: ["api", "v1", "plans"]}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.length > 0) pm.collectionVariables.set('planId', res.data[0].id);"]}}]},
                {name: "Get My Subscription", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/my", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "my"]}}},
                {name: "List Invoices", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/invoices", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "invoices"]}}},
                {name: "Get Invoice by ID", request: {method: "GET", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/invoices/{{invoiceId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "invoices", "{{invoiceId}}"]}}},
                {name: "Cancel Pending Invoice", request: {method: "DELETE", header: [{key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/invoices/{{invoiceId}}", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "invoices", "{{invoiceId}}"]}}},
                {name: "Checkout", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/checkout", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "checkout"]}, body: {mode: "raw", raw: "{\n  \"plan_id\": \"{{planId}}\",\n  \"seat_count\": 10,\n  \"billing_cycle\": \"monthly\",\n  \"payer_email\": \"owner@acme.com\"\n}"}}, event: [{listen: "test", script: {type: "text/javascript", exec: ["const res = pm.response.json();", "if (res && res.success && res.data && res.data.invoice && res.data.invoice.id) pm.collectionVariables.set('invoiceId', res.data.invoice.id);"]}}]},
                {name: "Upgrade Plan", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/upgrade", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "upgrade"]}, body: {mode: "raw", raw: "{\n  \"plan_id\": \"{{planId}}\",\n  \"seat_count\": 20,\n  \"payer_email\": \"owner@acme.com\"\n}"}}},
                {name: "Downgrade Plan", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/downgrade", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "downgrade"]}, body: {mode: "raw", raw: "{\n  \"plan_id\": \"{{planId}}\"\n}"}}},
                {name: "Cancel Subscription", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/cancel", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "cancel"]}, body: {mode: "raw", raw: "{\n  \"reason\": \"Switching to another provider\"\n}"}}},
                {name: "Change Seats", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}, {key: "Authorization", value: "Bearer {{accessToken}}"}], url: {raw: "{{baseUrl}}/api/v1/subscription/seats", host: ["{{baseUrl}}"], path: ["api", "v1", "subscription", "seats"]}, body: {mode: "raw", raw: "{\n  \"seat_count\": 15\n}"}}},
                {name: "Xendit Webhook (Public)", request: {method: "POST", header: [{key: "Content-Type", value: "application/json"}], url: {raw: "{{baseUrl}}/api/v1/webhook/xendit", host: ["{{baseUrl}}"], path: ["api", "v1", "webhook", "xendit"]}, body: {mode: "raw", raw: "{\n  \"id\": \"inv_123456\",\n  \"external_id\": \"sub-invoice-id\",\n  \"status\": \"PAID\",\n  \"amount\": 500000,\n  \"paid_amount\": 500000,\n  \"paid_at\": \"2026-06-19T10:00:00Z\",\n  \"payer_email\": \"owner@acme.com\",\n  \"payment_method\": \"BANK_TRANSFER\",\n  \"payment_channel\": \"BCA\"\n}"}}}
            ]
        }
    ]
};

fs.writeFileSync('postman_collection.json', JSON.stringify(collection, null, 2));
