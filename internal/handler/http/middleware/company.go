package middleware

import (
	"fmt"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/jwtauth/v5"
)

func RequireCompany(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			response.HandleError(w, fmt.Errorf("token claim can not be extracted"))
			return
		}

		companyID, companyFound := claims["company_id"].(string)
		if companyID == "" || !companyFound {
			response.HandleError(w, user.ErrCompanyIDRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}
