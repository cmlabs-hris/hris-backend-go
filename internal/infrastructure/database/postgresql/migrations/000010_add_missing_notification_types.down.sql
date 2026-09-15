ALTER TABLE notifications DROP CONSTRAINT IF EXISTS valid_notification_type;

DELETE FROM notifications WHERE type IN ('attendance_auto_closed', 'attendance_marked_absent');

ALTER TABLE notifications ADD CONSTRAINT valid_notification_type CHECK (
    type IN (
        'attendance_clock_in',
        'attendance_clock_out',
        'leave_request',
        'leave_approved',
        'leave_rejected',
        'payroll_generated',
        'schedule_updated',
        'invitation_sent',
        'employee_joined'
    )
);