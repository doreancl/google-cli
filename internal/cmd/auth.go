package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"google-cli/internal/authstore"

	"golang.org/x/oauth2"
)

var (
	loadOAuthConfigFn   = authstore.LoadOAuthConfig
	oauthTokenFromWebFn = oauthTokenFromWeb
	saveTokenFn         = authstore.SaveToken
	tokenPathFn         = authstore.TokenPath
	credentialsPathFn   = authstore.CredentialsPath
	openBrowserFn       = openBrowser
)

func runAuth(ctx context.Context, args []string) error {
	if len(args) > 0 && args[0] == "credentials" {
		return runAuthCredentials(ctx, args[1:])
	}
	if len(args) > 0 && args[0] == "add" {
		return runAuthAdd(ctx, args[1:])
	}

	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	credentialsPath := fs.String("credentials", "", "Ruta a client_secret.json")
	if err := fs.Parse(args); err != nil {
		return &ExitError{Code: 2, Err: err}
	}
	if fs.NArg() > 0 {
		return &ExitError{Code: 2, Err: fmt.Errorf("argumentos no soportados para auth: %s", strings.Join(fs.Args(), " "))}
	}

	cfg, err := loadOAuthConfigFn(*credentialsPath)
	if err != nil {
		return err
	}

	tok, err := oauthTokenFromWebFn(ctx, cfg)
	if err != nil {
		return err
	}
	if err := saveTokenFn(tok); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "Token guardado en %s\n", tokenPathFn())
	return nil
}

func runAuthAdd(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("auth add", flag.ContinueOnError)
	credentialsPath := fs.String("credentials", "", "Ruta a client_secret.json")
	if err := fs.Parse(args); err != nil {
		return &ExitError{Code: 2, Err: err}
	}
	if fs.NArg() != 1 {
		return &ExitError{Code: 2, Err: errors.New("uso: dorean_g auth add <email> [--credentials path]")}
	}
	email := strings.TrimSpace(fs.Arg(0))
	if email == "" || strings.Contains(email, " ") || !strings.Contains(email, "@") {
		return &ExitError{Code: 2, Err: fmt.Errorf("email invalido: %q", email)}
	}

	cfg, err := loadOAuthConfigFn(*credentialsPath)
	if err != nil {
		return err
	}
	tok, err := oauthTokenFromWebFn(ctx, cfg)
	if err != nil {
		return err
	}
	if err := saveTokenFn(tok); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, "Token guardado en %s para %s\n", tokenPathFn(), email)
	return nil
}

func runAuthCredentials(_ context.Context, args []string) error {
	if len(args) != 1 {
		return &ExitError{Code: 2, Err: errors.New("uso: dorean_g auth credentials <ruta.json> | dorean_g auth credentials list")}
	}

	if args[0] == "list" {
		return runAuthCredentialsList()
	}

	pathOverride := strings.TrimSpace(args[0])
	if pathOverride == "" {
		return &ExitError{Code: 2, Err: errors.New("ruta de credenciales vacia")}
	}
	if _, err := loadOAuthConfigFn(pathOverride); err != nil {
		return err
	}
	data, err := os.ReadFile(pathOverride) //nolint:gosec // user-provided path
	if err != nil {
		return fmt.Errorf("no pude leer credenciales en %s: %w", pathOverride, err)
	}
	path := credentialsPathFn()
	if mkErr := os.MkdirAll(filepath.Dir(path), 0o700); mkErr != nil {
		return mkErr
	}
	if writeErr := os.WriteFile(path, data, 0o600); writeErr != nil {
		return writeErr
	}
	_, _ = fmt.Fprintf(os.Stdout, "credentials_path\t%s\n", path)
	return nil
}

func runAuthCredentialsList() error {
	path := credentialsPathFn()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(os.Stderr, "No hay credenciales OAuth guardadas.")
			return nil
		}
		return fmt.Errorf("no pude leer credenciales en %s: %w", path, err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "credentials_path\t%s\n", path)
	return nil
}

func oauthTokenFromWeb(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	authURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	_, _ = fmt.Fprintf(os.Stdout, "Abre esta URL y pega el codigo:\n%s\n\n", authURL)
	_ = openBrowserFn(ctx, authURL)

	_, _ = fmt.Fprint(os.Stdout, "Codigo: ")
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}

	tok, err := cfg.Exchange(ctx, strings.TrimSpace(code))
	if err != nil {
		return nil, fmt.Errorf("no se pudo intercambiar el token: %w", err)
	}
	return tok, nil
}

func openBrowser(ctx context.Context, url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "windows":
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	}
	return cmd.Start()
}
