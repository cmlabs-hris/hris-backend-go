package middleware

import (
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/jwtauth/v5"
)

func AuthRequired(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			if err != nil {
				response.Unauthorized(w, err.Error())
				return
			}

			if token == nil {
				response.HandleError(w, auth.ErrInvalidToken)
				return
			}

			claims, err := token.AsMap(r.Context())
			if err != nil {
				response.HandleError(w, auth.ErrInvalidToken)
				return
			}
			tokenType, ok := claims["type"].(string)
			if tokenType != "access" || !ok {
				response.HandleError(w, auth.ErrInvalidToken)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
