-- Drop indexes
DROP INDEX IF EXISTS idx_audit_trails_user;
DROP INDEX IF EXISTS idx_audit_trails_record;
DROP INDEX IF EXISTS idx_employee_job_history_work_schedule_id;
DROP INDEX IF EXISTS idx_employee_job_history_branch_id;
DROP INDEX IF EXISTS idx_employee_job_history_grade_id;
DROP INDEX IF EXISTS idx_employee_job_history_position_id;
DROP INDEX IF EXISTS idx_employee_job_history_employee_id;
DROP INDEX IF EXISTS idx_employee_documents_employee_id;
DROP INDEX IF EXISTS idx_leave_requests_status;
DROP INDEX IF EXISTS idx_leave_requests_employee_id;
DROP INDEX IF EXISTS idx_attendances_employee_date;
DROP INDEX IF EXISTS idx_attendances_date;
DROP INDEX IF EXISTS idx_attendances_employee_id;
DROP INDEX IF EXISTS idx_employees_branch_id;
DROP INDEX IF EXISTS idx_employees_grade_id;
DROP INDEX IF EXISTS idx_employees_position_id;
DROP INDEX IF EXISTS idx_employees_user_id;
DROP INDEX IF EXISTS idx_employees_company_id;
DROP INDEX IF EXISTS idx_users_company_id;

-- Drop tables in dependency order
DROP TABLE IF EXISTS audit_trails;
DROP TABLE IF EXISTS employee_job_history;
DROP TABLE IF EXISTS employee_documents;
DROP TABLE IF EXISTS document_templates;
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS leave_quotas;
DROP TABLE IF EXISTS attendances;
DROP TABLE IF EXISTS employee_schedule_assignments;
DROP TABLE IF EXISTS work_schedule_locations;
DROP TABLE IF EXISTS work_schedule_times;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS document_types;
DROP TABLE IF EXISTS leave_types;
DROP TABLE IF EXISTS work_schedules;
DROP TABLE IF EXISTS branches;
DROP TABLE IF EXISTS grades;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS organization_units;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS companies;

-- Drop types
DROP TYPE IF EXISTS audit_action;
DROP TYPE IF EXISTS leave_request_status_enum;
DROP TYPE IF EXISTS employment_status_enum;
DROP TYPE IF EXISTS employment_type_enum;