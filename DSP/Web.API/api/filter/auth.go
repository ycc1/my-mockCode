package filter

import (
	"net/http"

	"advertiser-api/api/model"
	"advertiser-api/service"
)

func Authentication(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(service.SessionCookieName)
			if err != nil {
				model.ErrorResponse(w, http.StatusUnauthorized, "authentication required")
				return
			}
			username, valid := auth.Authenticate(cookie.Value)
			if !valid {
				model.ErrorResponse(w, http.StatusUnauthorized, "authentication required")
				return
			}
			request := r.WithContext(service.WithUsername(r.Context(), username))
			next.ServeHTTP(w, request)
		})
	}
}

func Attributes(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, authenticated := service.UsernameFromContext(r.Context()); !authenticated {
				model.ErrorResponse(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if required == "" {
				model.ErrorResponse(w, http.StatusForbidden, "API attribute is required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Offer(auth *service.AuthService, next http.Handler) http.Handler {
	return Authentication(auth)(Attributes("offer")(next))
}
