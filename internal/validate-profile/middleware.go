package validate_profile

import (
	"net/http"
	"strings"
)

// BearerTokenAuth is middleware for checking for the presence of a Bearer token in the Authorization header.
// Designed to be used with protected HTTP routes
func BearerTokenAuth(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if len(authHeader) == 0 {
				authFailed(w)
				return
			}

			tokens := strings.Split(authHeader, " ")
			if len(tokens) != 2 && tokens[0] != "Bearer" {
				authFailed(w)
				return
			}

			if tokens[1] != key {
				authFailed(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func authFailed(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", "Bearer")
	w.WriteHeader(http.StatusUnauthorized)
}
