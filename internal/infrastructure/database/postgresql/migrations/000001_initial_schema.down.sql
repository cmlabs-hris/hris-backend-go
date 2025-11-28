-- Drop indexes
-- PUBLIC_HOLIDAYS
DROP INDEX IF EXISTS idx_public_holidays_company_date_range;
DROP INDEX IF EXISTS idx_public_holidays_date;

-- AUDIT_TRAILS
DROP INDEX IF EXISTS idx_audit_trails_user_created;
DROP INDEX IF EXISTS idx_audit_trails_user;
DROP INDEX IF EXISTS idx_audit_trails_record;

-- EMPLOYEE_JOB_HISTORY
DROP INDEX IF EXISTS idx_employee_job_history_employee_dates;
DROP INDEX IF EXISTS idx_employee_job_history_work_schedule_id;
DROP INDEX IF EXISTS idx_employee_job_history_branch_id;
DROP INDEX IF EXISTS idx_employee_job_history_grade_id;
DROP INDEX IF EXISTS idx_employee_job_history_position_id;
DROP INDEX IF EXISTS idx_employee_job_history_employee_id;

-- EMPLOYEE_DOCUMENTS
DROP INDEX IF EXISTS idx_employee_documents_employee_id;

-- REFRESH_TOKENS
DROP INDEX IF EXISTS idx_refresh_tokens_user_expires;

-- EMPLOYEE_SCHEDULE_ASSIGNMENTS
DROP INDEX IF EXISTS idx_employee_schedule_assignments_date_range;
DROP INDEX IF EXISTS idx_employee_schedule_assignments_employee_dates;
DROP INDEX IF EXISTS idx_employee_schedule_assignments_employee_id;

-- WORK_SCHEDULE_LOCATIONS
DROP INDEX IF EXISTS idx_work_schedule_locations_schedule_id;

-- WORK_SCHEDULE_TIMES
DROP INDEX IF EXISTS idx_work_schedule_times_schedule_id;

-- WORK_SCHEDULES
DROP INDEX IF EXISTS idx_work_schedules_id_company_active;
DROP INDEX IF EXISTS idx_work_schedules_name_trgm;
DROP INDEX IF EXISTS idx_unique_schedule_name;
DROP INDEX IF EXISTS idx_work_schedules_deleted_at;
DROP INDEX IF EXISTS idx_work_schedules_type;
DROP INDEX IF EXISTS idx_work_schedules_company_id;

-- ATTENDANCES
DROP INDEX IF EXISTS idx_attendances_id_company;
DROP INDEX IF EXISTS idx_attendances_employee_open_session;
DROP INDEX IF EXISTS idx_attendances_company_id;
DROP INDEX IF EXISTS idx_attendances_employee_date;
DROP INDEX IF EXISTS idx_attendances_date;
DROP INDEX IF EXISTS idx_attendances_employee_id;

-- LEAVE_REQUESTS
DROP INDEX IF EXISTS idx_leave_requests_composite;
DROP INDEX IF EXISTS idx_leave_requests_submitted_at;
DROP INDEX IF EXISTS idx_leave_requests_employee_date_overlap;
DROP INDEX IF EXISTS idx_leave_requests_employee_leave_type;
DROP INDEX IF EXISTS idx_leave_requests_status;
DROP INDEX IF EXISTS idx_leave_requests_employee_id;
DROP INDEX IF EXISTS idx_leave_requests_date_range;
DROP INDEX IF EXISTS idx_leave_requests_employee_status;

-- LEAVE_QUOTAS
DROP INDEX IF EXISTS idx_leave_quotas_composite;
DROP INDEX IF EXISTS idx_leave_quotas_employee_year;
DROP INDEX IF EXISTS idx_leave_quotas_employee_id;

-- LEAVE_TYPES
DROP INDEX IF EXISTS idx_leave_types_company_active;
DROP INDEX IF EXISTS idx_leave_types_company_id;

-- EMPLOYEES
DROP INDEX IF EXISTS idx_employees_full_name_trgm;
DROP INDEX IF EXISTS idx_employees_company_status;
DROP INDEX IF EXISTS idx_employees_branch_id;
DROP INDEX IF EXISTS idx_employees_grade_id;
DROP INDEX IF EXISTS idx_employees_position_id;
DROP INDEX IF EXISTS idx_employees_user_id;
DROP INDEX IF EXISTS idx_employees_company_id;

-- USERS
DROP INDEX IF EXISTS idx_users_company_role;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_company_id;

-- Drop tables in correct reverse dependency order
DROP TABLE IF EXISTS audit_trails;
DROP TABLE IF EXISTS employee_job_history;
DROP TABLE IF EXISTS employee_documents;
DROP TABLE IF EXISTS document_templates;
DROP TABLE IF EXISTS document_types;
DROP TABLE IF EXISTS public_holidays;
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS leave_quotas;
DROP TABLE IF EXISTS attendances; -- Must drop before leave_types
DROP TABLE IF EXISTS leave_types;
DROP TABLE IF EXISTS employee_schedule_assignments;
DROP TABLE IF EXISTS work_schedule_locations;
DROP TABLE IF EXISTS work_schedule_times;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS work_schedules;
DROP TABLE IF EXISTS branches;
DROP TABLE IF EXISTS grades;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS companies;

-- Drop extension
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS btree_gist;

-- Drop types
DROP TYPE IF EXISTS leave_duration_enum;
DROP TYPE IF EXISTS leave_request_status_enum;
DROP TYPE IF EXISTS audit_action;
DROP TYPE IF EXISTS employment_status_enum;
DROP TYPE IF EXISTS employment_type_enum;