package interfaces

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const siteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type VerificationResult struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

func Validate(ctx context.Context, token, remoteIP string) (*VerificationResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("empty turnstile token")
	}

	secret := os.Getenv("TURNSTILE_SECRET_KEY")
	if secret == "" {
		return nil, errors.New("TURNSTILE_SECRET_KEY is not set")
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, siteverifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VerificationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			return &result, errors.New(strings.Join(result.ErrorCodes, ", "))
		}
		return &result, errors.New("turnstile validation failed")
	}

	return &result, nil
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
