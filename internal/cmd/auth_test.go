package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/oauth2"
)

func TestRunAuthParseError(t *testing.T) {
	err := runAuth(context.Background(), []string{"--unknown-flag"})
	if err == nil {
		t.Fatal("expected parse error")
	}
	var ee *ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected ExitError, got %T", err)
	}
}

func TestOAuthTokenFromWebSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-1",
			"refresh_token": "refresh-1",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer ts.Close()

	cfg := &oauth2.Config{
		ClientID:     "cid",
		ClientSecret: "sec",
		RedirectURL:  "http://localhost",
		Endpoint: oauth2.Endpoint{
			AuthURL:  ts.URL + "/auth",
			TokenURL: ts.URL + "/token",
		},
	}

	f, err := os.CreateTemp(t.TempDir(), "code-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fmt.Fprintln(f, "auth-code")
	if _, seekErr := f.Seek(0, 0); seekErr != nil {
		t.Fatal(seekErr)
	}
	orig := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = orig }()

	tok, err := oauthTokenFromWeb(context.Background(), cfg)
	if err != nil {
		t.Fatalf("oauthTokenFromWeb: %v", err)
	}
	if tok.AccessToken != "access-1" || tok.RefreshToken != "refresh-1" {
		t.Fatalf("unexpected token: %+v", tok)
	}
}

func TestRunAuthDependencyPaths(t *testing.T) {
	origLoad := loadOAuthConfigFn
	origOAuth := oauthTokenFromWebFn
	origSave := saveTokenFn
	origPath := tokenPathFn
	defer func() {
		loadOAuthConfigFn = origLoad
		oauthTokenFromWebFn = origOAuth
		saveTokenFn = origSave
		tokenPathFn = origPath
	}()

	loadOAuthConfigFn = func(string) (*oauth2.Config, error) { return nil, errors.New("load boom") }
	if err := runAuth(context.Background(), []string{}); err == nil {
		t.Fatal("expected load error")
	}

	loadOAuthConfigFn = func(string) (*oauth2.Config, error) { return &oauth2.Config{}, nil }
	oauthTokenFromWebFn = func(context.Context, *oauth2.Config) (*oauth2.Token, error) { return nil, errors.New("oauth boom") }
	if err := runAuth(context.Background(), []string{}); err == nil {
		t.Fatal("expected oauth error")
	}

	oauthTokenFromWebFn = func(context.Context, *oauth2.Config) (*oauth2.Token, error) {
		return &oauth2.Token{AccessToken: "a"}, nil
	}
	saveTokenFn = func(*oauth2.Token) error { return errors.New("save boom") }
	if err := runAuth(context.Background(), []string{}); err == nil {
		t.Fatal("expected save error")
	}

	saveTokenFn = func(*oauth2.Token) error { return nil }
	tokenPathFn = func() string { return "/tmp/token.json" }
	if err := runAuth(context.Background(), []string{}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
