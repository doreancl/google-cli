package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestRunAuthRejectsPositionalArgs(t *testing.T) {
	err := runAuth(context.Background(), []string{"unexpected"})
	if err == nil {
		t.Fatal("expected positional args error")
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected ExitError code 2, got %v", err)
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

func TestRunAuthCredentialsList(t *testing.T) {
	origCredPath := credentialsPathFn
	origLoad := loadOAuthConfigFn
	defer func() { credentialsPathFn = origCredPath }()
	defer func() { loadOAuthConfigFn = origLoad }()

	t.Run("usage error", func(t *testing.T) {
		err := runAuth(context.Background(), []string{"credentials"})
		if err == nil {
			t.Fatal("expected usage error")
		}
		var ee *ExitError
		if !errors.As(err, &ee) || ee.Code != 2 {
			t.Fatalf("expected ExitError code 2, got %v", err)
		}
	})

	t.Run("set credentials success", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "client_secret.json")
		content := []byte(`{"installed":{"client_id":"id"}}`)
		if err := os.WriteFile(src, content, 0o600); err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(t.TempDir(), "managed", "client_secret.json")
		credentialsPathFn = func() string { return dst }
		loadOAuthConfigFn = func(path string) (*oauth2.Config, error) {
			if path != src {
				t.Fatalf("expected src path, got %q", path)
			}
			return &oauth2.Config{}, nil
		}

		out := captureStdoutAuth(t, func() {
			if err := runAuth(context.Background(), []string{"credentials", src}); err != nil {
				t.Fatalf("runAuth credentials set: %v", err)
			}
		})
		if !strings.Contains(out, "credentials_path\t"+dst) {
			t.Fatalf("unexpected output: %q", out)
		}
		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(content) {
			t.Fatalf("content mismatch: got %q want %q", got, content)
		}
	})

	t.Run("set credentials invalid", func(t *testing.T) {
		loadOAuthConfigFn = func(string) (*oauth2.Config, error) { return nil, errors.New("invalid creds") }
		err := runAuth(context.Background(), []string{"credentials", "/tmp/invalid.json"})
		if err == nil || !strings.Contains(err.Error(), "invalid creds") {
			t.Fatalf("expected invalid creds error, got %v", err)
		}
	})

	t.Run("no credentials", func(t *testing.T) {
		credentialsPathFn = func() string { return filepath.Join(t.TempDir(), "client_secret.json") }
		stderr := captureStderrAuth(t, func() {
			if err := runAuth(context.Background(), []string{"credentials", "list"}); err != nil {
				t.Fatalf("runAuth credentials list: %v", err)
			}
		})
		if !strings.Contains(stderr, "No hay credenciales OAuth guardadas.") {
			t.Fatalf("unexpected stderr: %q", stderr)
		}
	})

	t.Run("credentials exist", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "client_secret.json")
		if err := os.WriteFile(p, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
		credentialsPathFn = func() string { return p }
		out := captureStdoutAuth(t, func() {
			if err := runAuth(context.Background(), []string{"credentials", "list"}); err != nil {
				t.Fatalf("runAuth credentials list: %v", err)
			}
		})
		if !strings.Contains(out, "credentials_path\t"+p) {
			t.Fatalf("unexpected output: %q", out)
		}
	})
}

func TestRunAuthCredentialsSetCommand(t *testing.T) {
	origCredPath := credentialsPathFn
	origLoad := loadOAuthConfigFn
	defer func() { credentialsPathFn = origCredPath }()
	defer func() { loadOAuthConfigFn = origLoad }()

	src := filepath.Join(t.TempDir(), "client_secret.json")
	content := []byte(`{"installed":{"client_id":"id"}}`)
	if err := os.WriteFile(src, content, 0o600); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "managed", "client_secret.json")
	credentialsPathFn = func() string { return dst }
	loadOAuthConfigFn = func(path string) (*oauth2.Config, error) {
		if path != src {
			t.Fatalf("expected src path, got %q", path)
		}
		return &oauth2.Config{}, nil
	}

	out := captureStdoutAuth(t, func() {
		if err := runAuth(context.Background(), []string{"credentials", src}); err != nil {
			t.Fatalf("runAuth credentials set: %v", err)
		}
	})
	if !strings.Contains(out, "credentials_path\t"+dst) {
		t.Fatalf("unexpected output: %q", out)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("content mismatch: got %q want %q", got, content)
	}
}

func TestRunAuthAdd(t *testing.T) {
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

	t.Run("usage", func(t *testing.T) {
		err := runAuth(context.Background(), []string{"add"})
		if err == nil {
			t.Fatal("expected usage error")
		}
		var ee *ExitError
		if !errors.As(err, &ee) || ee.Code != 2 {
			t.Fatalf("expected ExitError code 2, got %v", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		err := runAuth(context.Background(), []string{"add", "not-email"})
		if err == nil {
			t.Fatal("expected invalid email error")
		}
		var ee *ExitError
		if !errors.As(err, &ee) || ee.Code != 2 {
			t.Fatalf("expected ExitError code 2, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		loadOAuthConfigFn = func(string) (*oauth2.Config, error) { return &oauth2.Config{}, nil }
		oauthTokenFromWebFn = func(context.Context, *oauth2.Config) (*oauth2.Token, error) {
			return &oauth2.Token{AccessToken: "a"}, nil
		}
		saveTokenFn = func(*oauth2.Token) error { return nil }
		tokenPathFn = func() string { return "/tmp/token.json" }
		out := captureStdoutAuth(t, func() {
			if err := runAuth(context.Background(), []string{"add", "you@gmail.com"}); err != nil {
				t.Fatalf("runAuth add: %v", err)
			}
		})
		if !strings.Contains(out, "Token guardado en /tmp/token.json para you@gmail.com") {
			t.Fatalf("unexpected output: %q", out)
		}
	})
}

func captureStdoutAuth(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	if closeErr := w.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	b, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(b)
}

func captureStderrAuth(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()

	if closeErr := w.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	b, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(b)
}
