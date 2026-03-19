package console

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type adminSession struct {
	Username  string    `json:"username"`
	CSRFToken string    `json:"csrf_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func hashAdminPassword(password string) ([]byte, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return nil, fmt.Errorf("admin password must not be empty")
	}
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (s *Service) createAdminSession(ctx context.Context) (string, string, error) {
	sessionID := "admin_" + randomToken(32)
	csrfToken := "csrf_" + randomToken(32)
	session := adminSession{
		Username:  s.adminUsername,
		CSRFToken: csrfToken,
		ExpiresAt: time.Now().Add(s.sessionTTL),
	}
	if err := s.store.SaveAdminSession(ctx, sessionID, session, s.sessionTTL); err != nil {
		return "", "", err
	}
	return sessionID, csrfToken, nil
}

func (s *Service) validateAdminSession(ctx context.Context, sessionID string) (adminSession, bool, error) {
	if strings.TrimSpace(sessionID) == "" {
		return adminSession{}, false, nil
	}
	session, ok, err := s.store.LoadAdminSession(ctx, strings.TrimSpace(sessionID))
	if err != nil || !ok {
		return adminSession{}, ok, err
	}
	if session.ExpiresAt.Before(time.Now()) {
		_ = s.store.DeleteAdminSession(ctx, strings.TrimSpace(sessionID))
		return adminSession{}, false, nil
	}
	return session, true, nil
}

func (s *Service) deleteAdminSession(ctx context.Context, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}
	return s.store.DeleteAdminSession(ctx, strings.TrimSpace(sessionID))
}

func (s *Service) CreateAdminSession(ctx context.Context) (string, string, error) {
	return s.createAdminSession(ctx)
}

func (s *Service) ValidateAdminSession(ctx context.Context, sessionID string) (adminSession, bool, error) {
	return s.validateAdminSession(ctx, sessionID)
}

func (s *Service) DeleteAdminSession(ctx context.Context, sessionID string) error {
	return s.deleteAdminSession(ctx, sessionID)
}

func (s *Service) AuthenticateAdmin(username, password string) bool {
	if strings.TrimSpace(username) != s.adminUsername {
		return false
	}
	return bcrypt.CompareHashAndPassword(s.adminPasswordHash, []byte(password)) == nil
}
