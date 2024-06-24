package fal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	envAuthToken     = "FAL_AUTH_TOKEN"
	proxyUrl         = "https://fal.run/fal-ai/"
	defaultUserAgent = "fal/go"
	ErrNoAuth        = errors.New(`no auth token or token source provided`)
	ErrEnvVarNotSet  = fmt.Errorf("%s environment variable not set", envAuthToken)
	ErrEnvVarEmpty   = fmt.Errorf("%s environment variable is empty", envAuthToken)
)

type Client struct {
	options *clientOptions
	c       *http.Client
}

type clientOptions struct {
	auth       string
	baseUrl    string
	httpClient *http.Client
	userAgent  string
	subdomain  string
}

type ClientOption func(*clientOptions) error

func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		options: &clientOptions{
			httpClient: http.DefaultClient,
			userAgent:  defaultUserAgent,
			baseUrl:    proxyUrl,
		},
	}
	var errs []error

	for _, opt := range opts {
		if err := opt(c.options); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		if err != nil {
			return nil, err
		}
		return nil, errors.New("failed to apply options")
	}

	if c.options.auth == "" {
		return nil, ErrNoAuth
	}
	c.c = c.options.httpClient
	return c, nil
}

func WithToken(token string) ClientOption {
	return func(o *clientOptions) error {
		o.auth = token
		return nil
	}
}

func WithTokenFromEnv() ClientOption {
	return func(o *clientOptions) error {
		token, ok := os.LookupEnv(envAuthToken)
		if !ok {
			return ErrEnvVarEmpty
		}

		if token == "" {
			return ErrEnvVarNotSet
		}
		o.auth = token
		return nil
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(o *clientOptions) error {
		o.userAgent = userAgent
		return nil
	}
}

func WithHttpClient(httpClient *http.Client) ClientOption {
	return func(o *clientOptions) error {
		o.httpClient = httpClient
		return nil
	}
}

func WithSubdomain(subdomain string) ClientOption {
	return func(o *clientOptions) error {
		o.baseUrl = fmt.Sprintf("https://%s.fal.run/fal-ai/", subdomain)
		return nil
	}
}

func constructUrl(baseUrl, route string) string {
	route = strings.TrimPrefix(route, "/")

	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}

	return baseUrl + route
}

func (r *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := constructUrl(r.options.baseUrl, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", r.options.auth))
	if r.options.userAgent != "" {
		req.Header.Set("User-Agent", r.options.userAgent)
	}

	return req, nil
}

func (r *Client) Fetch(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	bodyBuffer := &bytes.Buffer{}

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyBuffer = bytes.NewBuffer(bodyBytes)
	}
	req, err := r.newRequest(ctx, method, path, bodyBuffer)
	if err != nil {
		return err
	}

	return r.do(req, out)
}

func (r *Client) do(request *http.Request, out interface{}) error {
	resp, err := r.c.Do(request)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to make request with status code: %v", resp.StatusCode)
	}

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if out != nil {
		if err := json.Unmarshal(responseBytes, &out); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
