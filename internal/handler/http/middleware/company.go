package middleware

import (
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/jwtauth/v5"
)

func RequireCompany(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			response.HandleError(w, user.ErrCompanyIDRequired)
		}

		companyID, companyFound := claims["company_id"].(*string)
		if err != nil {
			response.HandleError(w, user.ErrCompanyIDRequired)
		}

		role, roleFound := claims["role"].(user.Role)
		if err != nil {
			response.HandleError(w, user.ErrCompanyIDRequired)
		}

		if role == user.RolePending && roleFound {
			if companyID == nil || *companyID == "" || companyFound {
				response.HandleError(w, user.ErrCompanyIDRequired)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
