package googleapi

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"google-cli/internal/authstore"

	"google.golang.org/api/option"
)

func writeOAuthFixture(t *testing.T) {
	t.Helper()
	cfgDir := authstore.ConfigDir()
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	creds := `{"installed":{"client_id":"id","project_id":"p","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","client_secret":"sec","redirect_uris":["http://localhost"]}}`
	if err := os.WriteFile(authstore.CredentialsPath(), []byte(creds), 0o600); err != nil {
		t.Fatal(err)
	}
	tok := map[string]string{"access_token": "a", "token_type": "Bearer", "refresh_token": "r"}
	b, _ := json.Marshal(tok)
	if err := os.WriteFile(authstore.TokenPath(), b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNewHTTPClientErrorsWhenMissingConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	_, err := newHTTPClient(context.Background())
	if err == nil {
		t.Fatal("expected missing credentials error")
	}
}

func TestNewHTTPClientSuccess(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	writeOAuthFixture(t)

	c, err := newHTTPClient(context.Background())
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	if c.Timeout != defaultTimeout {
		t.Fatalf("unexpected timeout: %v", c.Timeout)
	}
	if _, ok := c.Transport.(*oauthRetryTransport); !ok {
		t.Fatalf("expected oauthRetryTransport, got %T", c.Transport)
	}
}

func TestNewService(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	writeOAuthFixture(t)

	type stub struct{ N int }
	out, err := newService(context.Background(), func(context.Context, ...option.ClientOption) (*stub, error) {
		return &stub{N: 7}, nil
	})
	if err != nil {
		t.Fatalf("newService success: %v", err)
	}
	if out.N != 7 {
		t.Fatalf("unexpected value: %+v", out)
	}

	_, err = newService(context.Background(), func(context.Context, ...option.ClientOption) (*stub, error) {
		return nil, errors.New("factory boom")
	})
	if err == nil {
		t.Fatal("expected factory error")
	}
}
