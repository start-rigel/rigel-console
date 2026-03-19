package console

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type challengeVerifier interface {
	Available() bool
	Verify(ctx context.Context, token, clientIP string) error
}

type noopChallengeVerifier struct{}

func (noopChallengeVerifier) Available() bool { return false }
func (noopChallengeVerifier) Verify(context.Context, string, string) error {
	return fmt.Errorf("challenge verifier is not configured")
}

type turnstileChallengeVerifier struct {
	verifyURL string
	secret    string
	client    *http.Client
}

func NewChallengeVerifier(provider, verifyURL, secret string) challengeVerifier {
	if !strings.EqualFold(strings.TrimSpace(provider), "turnstile") || strings.TrimSpace(secret) == "" {
		return noopChallengeVerifier{}
	}
	return &turnstileChallengeVerifier{
		verifyURL: strings.TrimSpace(verifyURL),
		secret:    strings.TrimSpace(secret),
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (v *turnstileChallengeVerifier) Available() bool { return true }

func (v *turnstileChallengeVerifier) Verify(ctx context.Context, token, clientIP string) error {
	form := url.Values{}
	form.Set("secret", v.secret)
	form.Set("response", strings.TrimSpace(token))
	if strings.TrimSpace(clientIP) != "" {
		form.Set("remoteip", strings.TrimSpace(clientIP))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create challenge verify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("verify challenge: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("challenge provider returned %d", resp.StatusCode)
	}

	var payload struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode challenge response: %w", err)
	}
	if !payload.Success {
		return fmt.Errorf("challenge verification failed")
	}
	return nil
}
