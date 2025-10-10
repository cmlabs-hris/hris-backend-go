package middleware

import (
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/jwtauth/v5"
)

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			response.HandleError(w, auth.ErrInvalidToken)
			return
		}

		admin, ok := claims["is_admin"].(bool)
		if !admin || !ok {
			response.HandleError(w, user.ErrAdminPrivilegeRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}
