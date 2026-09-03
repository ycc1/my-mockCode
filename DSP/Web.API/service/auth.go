package service

import (
	"context"
	"net/http"

	"advertiser-api/repository"
)

const SessionCookieName = "advertiser_session"

type usernameContextKey struct{}

func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameContextKey{}).(string)
	return username, ok && username != ""
}

func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameContextKey{}, username)
}

type AuthService struct {
	credentials repository.CredentialRepository
	sessions    repository.SessionRepository
}

func NewAuthService(credentials repository.CredentialRepository, sessions repository.SessionRepository) *AuthService {
	return &AuthService{credentials: credentials, sessions: sessions}
}

func (s *AuthService) Login(username, password string) (string, bool, error) {
	valid, err := s.credentials.Validate(username, password)
	if err != nil {
		return "", false, err
	}
	if !valid {
		return "", false, nil
	}
	token, err := s.sessions.Create(username)
	if err != nil {
		return "", false, err
	}
	return token, true, nil
}

func (s *AuthService) Logout(token string) {
	if token != "" {
		s.sessions.Delete(token)
	}
}

func (s *AuthService) Authenticate(token string) (string, bool) {
	if token == "" {
		return "", false
	}
	return s.sessions.Get(token)
}

func (s *AuthService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		username, valid := "", false
		if err == nil {
			username, valid = s.Authenticate(cookie.Value)
		}
		if !valid {
			writeUnauthorized(w)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUsername(r.Context(), username)))
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"code":401,"message":"authentication required"}` + "\n"))
}
