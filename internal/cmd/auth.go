package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	openBrowserFn       = openBrowser
)

func runAuth(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	credentialsPath := fs.String("credentials", "", "Ruta a client_secret.json")
	if err := fs.Parse(args); err != nil {
		return &ExitError{Code: 2, Err: err}
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
