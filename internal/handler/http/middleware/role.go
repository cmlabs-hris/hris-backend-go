package middleware

import (
	"fmt"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/jwtauth/v5"
)

// RequireOwner requires owner role
func RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			response.HandleError(w, user.ErrOwnerAccessRequired)
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			response.HandleError(w, user.ErrOwnerAccessRequired)
			return
		}

		if role != string(user.RoleOwner) {
			response.HandleError(w, user.ErrOwnerAccessRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireManager requires manager or owner role
func RequireManager(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			response.HandleError(w, user.ErrManagerAccessRequired)
			return
		}

		roleStr, ok := claims["role"].(string)
		if !ok {
			response.HandleError(w, user.ErrManagerAccessRequired)
			return
		}

		role := user.Role(roleStr)
		if role != user.RoleManager && role != user.RoleOwner {
			response.HandleError(w, user.ErrManagerAccessRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequirePending(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			response.HandleError(w, user.ErrPendingRoleAccessRequired)
			return
		}

		roleStr, ok := claims["role"].(string)
		if !ok {
			response.HandleError(w, user.ErrPendingRoleAccessRequired)
			return
		}

		role := user.Role(roleStr)
		if role != user.RolePending {
			response.HandleError(w, user.ErrPendingRoleAccessRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks if user has specific permission
func RequirePermission(permission user.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				response.Forbidden(w, fmt.Sprintf("Insufficient permissions: required '%s'", permission))
				return
			}

			roleStr, ok := claims["role"].(string)
			if !ok {
				response.Forbidden(w, fmt.Sprintf("Insufficient permissions: required '%s'", permission))
				return
			}

			role := user.Role(roleStr)
			if !user.HasPermission(role, permission) {
				response.Forbidden(w, fmt.Sprintf("Insufficient permissions: required '%s', but user role is '%s'", permission, role))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
