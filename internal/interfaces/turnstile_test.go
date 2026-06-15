package interfaces

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTurnstileSiteKeyUsesDevelopmentKeyLocally(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("TURNSTILE_SITE_KEY", "production-site-key")

	siteKey, err := TurnstileSiteKey()
	if err != nil {
		t.Fatalf("TurnstileSiteKey() error = %v", err)
	}
	if siteKey != turnstileTestSiteKey {
		t.Fatalf("TurnstileSiteKey() = %q, want development test key", siteKey)
	}
}

func TestTurnstileSiteKeyRequiresProductionConfiguration(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("TURNSTILE_SITE_KEY", "")

	if _, err := TurnstileSiteKey(); err == nil {
		t.Fatal("TurnstileSiteKey() expected an error for missing production key")
	}
}

func TestValidateChecksActionInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("TURNSTILE_SECRET_KEY", "production-secret")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("secret"); got != "production-secret" {
			t.Errorf("secret = %q, want production secret", got)
		}
		if got := r.Form.Get("response"); got != "valid-token" {
			t.Errorf("response = %q, want valid-token", got)
		}
		if got := r.Form.Get("remoteip"); got != "127.0.0.1" {
			t.Errorf("remoteip = %q, want 127.0.0.1", got)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"success":true,"hostname":"localhost","action":"register"}`)
	}))
	defer server.Close()

	originalURL := turnstileVerifyURL
	originalClient := turnstileClient
	turnstileVerifyURL = server.URL
	turnstileClient = &http.Client{Timeout: time.Second}
	t.Cleanup(func() {
		turnstileVerifyURL = originalURL
		turnstileClient = originalClient
	})

	result, err := Validate(
		context.Background(),
		"valid-token",
		"127.0.0.1",
		"register",
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !result.Success {
		t.Fatal("Validate() returned unsuccessful result")
	}

	if _, err := Validate(
		context.Background(),
		"valid-token",
		"127.0.0.1",
		"login",
	); err == nil {
		t.Fatal("Validate() expected an error for mismatched action")
	}
}

func TestValidateAllowsDevelopmentResponseWithoutAction(t *testing.T) {
	t.Setenv("APP_ENV", "development")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("secret"); got != turnstileTestSecretKey {
			t.Errorf("secret = %q, want development test secret", got)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"success":true,"hostname":"example.com"}`)
	}))
	defer server.Close()

	originalURL := turnstileVerifyURL
	originalClient := turnstileClient
	turnstileVerifyURL = server.URL
	turnstileClient = &http.Client{Timeout: time.Second}
	t.Cleanup(func() {
		turnstileVerifyURL = originalURL
		turnstileClient = originalClient
	})

	if _, err := Validate(
		context.Background(),
		"XXXX.DUMMY.TOKEN.XXXX",
		"127.0.0.1",
		"register",
	); err != nil {
		t.Fatalf("Validate() development error = %v", err)
	}
}

func TestValidateRejectsOversizedToken(t *testing.T) {
	t.Setenv("APP_ENV", "development")

	_, err := Validate(
		context.Background(),
		strings.Repeat("x", maxTurnstileTokenLength+1),
		"",
		"login",
	)
	if err == nil {
		t.Fatal("Validate() expected an error for oversized token")
	}
}
