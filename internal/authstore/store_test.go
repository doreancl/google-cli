package authstore

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

func TestResolveCredentialsPathPriority(t *testing.T) {
	t.Setenv("GCAL_CREDENTIALS", "/env/creds.json")
	if got := ResolveCredentialsPath("/flag/creds.json"); got != "/flag/creds.json" {
		t.Fatalf("expected override path, got %q", got)
	}
	if got := ResolveCredentialsPath(""); got != "/env/creds.json" {
		t.Fatalf("expected env path, got %q", got)
	}
}

func TestConfigDirTokenPathAndDefaultCredentials(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("GCAL_CREDENTIALS", "")

	cfg := ConfigDir()
	wantCfg := filepath.Join(tmp, "Library", "Application Support", AppName)
	if cfg != wantCfg {
		t.Fatalf("ConfigDir mismatch: got %q want %q", cfg, wantCfg)
	}

	if got := TokenPath(); got != filepath.Join(wantCfg, "token.json") {
		t.Fatalf("TokenPath mismatch: %q", got)
	}

	if got := ResolveCredentialsPath(""); got != filepath.Join(wantCfg, "client_secret.json") {
		t.Fatalf("ResolveCredentialsPath default mismatch: %q", got)
	}
}

func TestLoadOAuthConfigReadError(t *testing.T) {
	_, err := LoadOAuthConfig("/no/existe/credenciales.json")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadOAuthConfigInvalidJSON(t *testing.T) {
	f := filepath.Join(t.TempDir(), "client_secret.json")
	if err := os.WriteFile(f, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadOAuthConfig(f)
	if err == nil {
		t.Fatal("expected invalid credentials error")
	}
}

func TestSaveAndReadToken(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")

	in := &oauth2.Token{AccessToken: "abc", TokenType: "Bearer", RefreshToken: "r1"}
	if err := SaveToken(in); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	out, err := ReadToken()
	if err != nil {
		t.Fatalf("ReadToken: %v", err)
	}
	if out.AccessToken != in.AccessToken || out.RefreshToken != in.RefreshToken || out.TokenType != in.TokenType {
		t.Fatalf("token mismatch: %+v vs %+v", out, in)
	}
}

func TestReadTokenMissingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	_, err := ReadToken()
	if err == nil {
		t.Fatal("expected missing token error")
	}
}
