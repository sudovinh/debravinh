package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRoutes(t *testing.T) {
	e := newServer()

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantLocation string
		wantBody     string
	}{
		{name: "home page", path: "/", wantStatus: http.StatusOK, wantBody: "Debra"},
		{name: "about us page", path: "/aboutus", wantStatus: http.StatusOK, wantBody: "Debra"},
		{name: "old wedding path redirects home", path: "/wedding", wantStatus: http.StatusTemporaryRedirect, wantLocation: "/"},
		{name: "levi redirects to 529 gifting", path: "/levi", wantStatus: http.StatusTemporaryRedirect, wantLocation: "https://digital.fidelity.com/prgw/digital/familygifting/mlgLandingPage?uuid=0ec1277a30464ccbbc25f69b04ca429a"},
		{name: "robots.txt", path: "/robots.txt", wantStatus: http.StatusOK, wantBody: "User-agent"},
		{name: "static asset", path: "/assets/css/style.css", wantStatus: http.StatusOK},
		{name: "unknown path redirects home", path: "/nonexistent", wantStatus: http.StatusTemporaryRedirect, wantLocation: "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GET %s status = %d, want %d", tt.path, rec.Code, tt.wantStatus)
			}
			if tt.wantLocation != "" {
				if got := rec.Header().Get("Location"); got != tt.wantLocation {
					t.Errorf("GET %s Location = %q, want %q", tt.path, got, tt.wantLocation)
				}
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("GET %s body does not contain %q", tt.path, tt.wantBody)
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	e := newServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Simulate the TLS-terminating proxy in front of the app (DigitalOcean App
	// Platform); HSTS is only sent on HTTPS requests, as the spec requires.
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	wantHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Strict-Transport-Security": "max-age=31536000; includeSubdomains",
	}
	for header, want := range wantHeaders {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("header %s = %q, want %q", header, got, want)
		}
	}

	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("Content-Security-Policy header is missing")
	}
	for _, directive := range []string{"default-src 'self'", "fonts.googleapis.com", "fonts.gstatic.com"} {
		if !strings.Contains(csp, directive) {
			t.Errorf("Content-Security-Policy missing %q, got %q", directive, csp)
		}
	}

	if got := rec.Header().Get("Permissions-Policy"); got == "" {
		t.Error("Permissions-Policy header is missing")
	}
}

func TestAssetCaching(t *testing.T) {
	e := newServer()

	req := httptest.NewRequest(http.MethodGet, "/assets/css/style.css", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); !strings.Contains(got, "max-age") {
		t.Errorf("Cache-Control on asset = %q, want a max-age directive", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); strings.Contains(got, "max-age=86400") {
		t.Errorf("Cache-Control on page = %q, pages should not be cached like assets", got)
	}
}

func TestGzipCompression(t *testing.T) {
	e := newServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Errorf("Content-Encoding = %q, want gzip", got)
	}
}
