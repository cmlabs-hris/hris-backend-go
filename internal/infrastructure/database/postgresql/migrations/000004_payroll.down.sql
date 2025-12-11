-- Rollback payroll schema
DROP INDEX IF EXISTS idx_payroll_records_employee_period;
DROP INDEX IF EXISTS idx_payroll_records_status;
DROP INDEX IF EXISTS idx_payroll_records_period;
DROP INDEX IF EXISTS idx_payroll_records_employee;
DROP INDEX IF EXISTS idx_payroll_records_company;
DROP INDEX IF EXISTS idx_emp_payroll_comp_active;
DROP INDEX IF EXISTS idx_emp_payroll_comp_component;
DROP INDEX IF EXISTS idx_emp_payroll_comp_employee;
DROP INDEX IF EXISTS idx_payroll_components_active;
DROP INDEX IF EXISTS idx_payroll_components_type;
DROP INDEX IF EXISTS idx_payroll_components_company;
DROP INDEX IF EXISTS idx_payroll_settings_company;

DROP TABLE IF EXISTS payroll_records;
DROP TABLE IF EXISTS employee_payroll_components;
DROP TABLE IF EXISTS payroll_components;
DROP TABLE IF EXISTS payroll_settings;

DROP TYPE IF EXISTS payroll_status;
DROP TYPE IF EXISTS payroll_component_type;

ALTER TABLE employees DROP COLUMN IF EXISTS base_salary;
