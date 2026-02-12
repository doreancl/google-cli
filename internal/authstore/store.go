package authstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	AppName       = "dorean_g"
	CalendarScope = "https://www.googleapis.com/auth/calendar"
)

func ConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return filepath.Join(dir, AppName)
}

func ResolveCredentialsPath(override string) string {
	if override != "" {
		return override
	}
	if env := os.Getenv("GCAL_CREDENTIALS"); env != "" {
		return env
	}
	return CredentialsPath()
}

func CredentialsPath() string {
	return filepath.Join(ConfigDir(), "client_secret.json")
}

func TokenPath() string {
	return filepath.Join(ConfigDir(), "token.json")
}

func LoadOAuthConfig(pathOverride string) (*oauth2.Config, error) {
	path := ResolveCredentialsPath(pathOverride)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no pude leer credenciales en %s: %w", path, err)
	}
	cfg, err := google.ConfigFromJSON(data, CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("credenciales invalidas: %w", err)
	}
	return cfg, nil
}

func SaveToken(token *oauth2.Token) error {
	path := TokenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func ReadToken() (*oauth2.Token, error) {
	f, err := os.Open(TokenPath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, err
	}
	return tok, nil
}
