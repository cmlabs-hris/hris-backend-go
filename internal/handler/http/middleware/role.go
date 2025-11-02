package middleware

import (
	"fmt"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
)

// RequireOwner requires owner role
func RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value("user_role").(user.Role)
		// if !ok {
		// 	response.Forbidden(w, "role required")
		// 	return
		// }

		if role != user.RoleOwner {
			response.HandleError(w, user.ErrOwnerAccessRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireManager requires manager or owner role
func RequireManager(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value("user_role").(user.Role)
		// if !ok {
		// 	response.Forbidden(w, "role required")
		// 	return
		// }

		if role != user.RoleManager && role != user.RoleOwner {
			response.HandleError(w, user.ErrManagerAccessRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequirePending(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("user_role").(user.Role)
		if role != user.RolePending || !ok {
			response.HandleError(w, user.ErrAdminPrivilegeRequired)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks if user has specific permission
func RequirePermission(permission user.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value("user_role").(user.Role)

			if !user.HasPermission(role, permission) {
				response.Forbidden(w, fmt.Sprintf("Insufficient permissions: required '%s', but user role is '%s'", permission, role))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
