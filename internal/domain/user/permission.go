package user

type Permission string

const (
	// Self Management
	PermissionViewOwnProfile Permission = "profile.view_own"
	PermissionEditOwnProfile Permission = "profile.edit_own"

	// Leave Management
	PermissionLeaveViewOwn     Permission = "leave.view_own"
	PermissionLeaveCreate      Permission = "leave.create"
	PermissionLeaveViewAll     Permission = "leave.view_all"
	PermissionLeaveApprove     Permission = "leave.approve"
	PermissionLeaveManageTypes Permission = "leave.manage_types"

	// Attendance Management
	PermissionAttendanceViewOwn Permission = "attendance.view_own"
	PermissionAttendanceCreate  Permission = "attendance.create"
	PermissionAttendanceViewAll Permission = "attendance.view_all"
	PermissionAttendanceApprove Permission = "attendance.approve"

	// Employee Management
	PermissionEmployeeViewAll Permission = "employee.view_all"
	PermissionEmployeeManage  Permission = "employee.manage"

	// Company Management
	PermissionCompanyView   Permission = "company.view"
	PermissionCompanyManage Permission = "company.manage"

	// Reports
	PermissionReportsView Permission = "reports.view"

	// User Management
	PermissionUserManage Permission = "user.manage"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[Role][]Permission{
	RoleOwner: {
		// Owner has all permissions
		PermissionViewOwnProfile,
		PermissionEditOwnProfile,
		PermissionLeaveViewOwn,
		PermissionLeaveCreate,
		PermissionLeaveViewAll,
		PermissionLeaveApprove,
		PermissionLeaveManageTypes,
		PermissionAttendanceViewOwn,
		PermissionAttendanceCreate,
		PermissionAttendanceViewAll,
		PermissionAttendanceApprove,
		PermissionEmployeeViewAll,
		PermissionEmployeeManage,
		PermissionCompanyView,
		PermissionCompanyManage,
		PermissionReportsView,
		PermissionUserManage,
	},
	RoleManager: {
		// Manager can approve and view team data
		PermissionViewOwnProfile,
		PermissionEditOwnProfile,
		PermissionLeaveViewOwn,
		PermissionLeaveCreate,
		PermissionLeaveViewAll,
		PermissionLeaveApprove,
		PermissionAttendanceViewOwn,
		PermissionAttendanceCreate,
		PermissionAttendanceViewAll,
		PermissionAttendanceApprove,
		PermissionEmployeeViewAll,
		PermissionCompanyView,
		PermissionReportsView,
	},
	RoleEmployee: {
		// Employee has basic access
		PermissionViewOwnProfile,
		PermissionEditOwnProfile,
		PermissionLeaveViewOwn,
		PermissionLeaveCreate,
		PermissionAttendanceViewOwn,
		PermissionAttendanceCreate,
	},
	RolePending: {
		// Pending role has no permissions
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, permission Permission) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}
