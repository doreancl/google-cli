package googleapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"google-cli/internal/authstore"

	"google.golang.org/api/option"
)

const defaultTimeout = 30 * time.Second

type serviceFactory[T any] func(context.Context, ...option.ClientOption) (*T, error)

func newService[T any](ctx context.Context, factory serviceFactory[T]) (*T, error) {
	httpClient, err := newHTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := factory(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create google service: %w", err)
	}
	return svc, nil
}

func newHTTPClient(ctx context.Context) (*http.Client, error) {
	cfg, err := authstore.LoadOAuthConfig("")
	if err != nil {
		return nil, err
	}
	tok, err := authstore.ReadToken()
	if err != nil {
		return nil, fmt.Errorf("token no encontrado, corre `dorean_g auth`: %w", err)
	}

	base := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}
	return &http.Client{
		Transport: &oauthRetryTransport{
			base:         cfg.Client(ctx, tok).Transport,
			fallbackBase: base,
		},
		Timeout: defaultTimeout,
	}, nil
}
