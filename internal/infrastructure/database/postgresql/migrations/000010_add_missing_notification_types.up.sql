-- Fix: add missing notification types used by attendance cron jobs
-- (attendance_auto_closed, attendance_marked_absent) to the check constraint.
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS valid_notification_type;

ALTER TABLE notifications ADD CONSTRAINT valid_notification_type CHECK (
    type IN (
        'attendance_clock_in',
        'attendance_clock_out',
        'attendance_auto_closed',
        'attendance_marked_absent',
        'leave_request',
        'leave_approved',
        'leave_rejected',
        'payroll_generated',
        'schedule_updated',
        'invitation_sent',
        'employee_joined'
    )
);