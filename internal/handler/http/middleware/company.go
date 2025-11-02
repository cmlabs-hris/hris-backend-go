package middleware

import (
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
)

func RequireCompany(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		companyID, companyFound := r.Context().Value("company_id").(*string)
		role, roleFound := r.Context().Value("role").(user.Role)
		if role == user.RolePending && roleFound {
			if companyID == nil || *companyID == "" || companyFound {
				response.HandleError(w, user.ErrCompanyIDRequired)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
