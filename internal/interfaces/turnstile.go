package interfaces

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	siteverifyURL           = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	turnstileTestSiteKey    = "1x00000000000000000000AA"
	turnstileTestSecretKey  = "1x0000000000000000000000000000000AA"
	maxTurnstileTokenLength = 2048
)

var (
	turnstileVerifyURL = siteverifyURL
	turnstileClient    = &http.Client{Timeout: 10 * time.Second}
)

type VerificationResult struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

func TurnstileSiteKey() (string, error) {
	if isDevelopmentEnvironment() {
		return turnstileTestSiteKey, nil
	}

	siteKey := strings.TrimSpace(os.Getenv("TURNSTILE_SITE_KEY"))
	if siteKey == "" {
		return "", errors.New("переменная TURNSTILE_SITE_KEY не задана")
	}
	return siteKey, nil
}

func Validate(ctx context.Context, token, remoteIP, expectedAction string) (*VerificationResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("токен Turnstile отсутствует")
	}
	if len(token) > maxTurnstileTokenLength {
		return nil, errors.New("токен Turnstile слишком длинный")
	}

	secret, err := turnstileSecretKey()
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, turnstileVerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := turnstileClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Turnstile вернул HTTP %d", resp.StatusCode)
	}

	var result VerificationResult
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			return &result, errors.New(strings.Join(result.ErrorCodes, ", "))
		}
		return &result, errors.New("проверка Turnstile не пройдена")
	}
	if !isDevelopmentEnvironment() && expectedAction != "" && result.Action != expectedAction {
		return &result, fmt.Errorf(
			"получено действие Turnstile %q, ожидалось %q",
			result.Action,
			expectedAction,
		)
	}

	return &result, nil
}

func turnstileSecretKey() (string, error) {
	if isDevelopmentEnvironment() {
		return turnstileTestSecretKey, nil
	}

	secret := strings.TrimSpace(os.Getenv("TURNSTILE_SECRET_KEY"))
	if secret == "" {
		return "", errors.New("переменная TURNSTILE_SECRET_KEY не задана")
	}
	return secret, nil
}

func isDevelopmentEnvironment() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) {
	case "development", "dev", "local", "test":
		return true
	default:
		return false
	}
}

func ClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}

	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
