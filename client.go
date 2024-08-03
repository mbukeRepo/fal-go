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
	"time"
)

var (
	envAuthToken       = "FAL_AUTH_TOKEN"
	proxyUrl           = "https://fal.run/"
	defaultUserAgent   = "fal/go"
	ErrNoAuth          = errors.New(`no auth token or token source provided`)
	ErrEnvVarNotSet    = fmt.Errorf("%s environment variable not set", envAuthToken)
	ErrEnvVarEmpty     = fmt.Errorf("%s environment variable is empty", envAuthToken)
	defaultRetryPolicy = &retryPolicy{
		maxRetries: 3,
		backoff: &ExponentialBackOff{
			Base:       1 * time.Second,
			Jitter:     100 * time.Millisecond,
			Multiplier: 2,
		},
	}
)

// Client is a client for the FAL API.
type Client struct {
	options *clientOptions
	c       *http.Client
	Queue   *Queue // for running long running tasks
}

type retryPolicy struct {
	maxRetries int
	backoff    Backoff
}

type clientOptions struct {
	auth        string
	baseUrl     string
	httpClient  *http.Client
	userAgent   string
	retryPolicy *retryPolicy
}

// ClientOption is a function that modifies an options struct.
type ClientOption func(*clientOptions) error

// NewClient creates a new FAL API client.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		options: &clientOptions{
			httpClient:  http.DefaultClient,
			userAgent:   defaultUserAgent,
			baseUrl:     proxyUrl,
			retryPolicy: defaultRetryPolicy,
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
	c.Queue = &Queue{
		c:         c,
		Subdomain: "queue",
	}
	return c, nil
}

// WithToken sets the auth token used by the client.
func WithToken(token string) ClientOption {
	return func(o *clientOptions) error {
		o.auth = token
		return nil
	}
}

// WithTokenFromEnv configures the client to use the auth token provided in the
// FAL_AUTH_TOKEN environment variable.
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

// WithBaseURL sets the base URL for the client.
func WithUserAgent(userAgent string) ClientOption {
	return func(o *clientOptions) error {
		o.userAgent = userAgent
		return nil
	}
}

// WithHTTPClient sets the HTTP client used by the client.
func WithHttpClient(httpClient *http.Client) ClientOption {
	return func(o *clientOptions) error {
		o.httpClient = httpClient
		return nil
	}
}

// WithRetryPolicy sets the retry policy used by the client.
func WithRetryPolicy(maxRetries int, backoff Backoff) ClientOption {
	return func(o *clientOptions) error {
		o.retryPolicy = &retryPolicy{
			maxRetries: maxRetries,
			backoff:    backoff,
		}
		return nil
	}
}

type QueryParams map[string]string

type UrlOptions struct {
	Subdomain string
	Query     *QueryParams
	AppId     string
}

func constructUrl(baseUrl, route string, urlOptions *UrlOptions) string {
	route = strings.TrimPrefix(route, "/")

	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}

	var queryParams string
	if urlOptions != nil && urlOptions.Query != nil {
		queryParams = "?"
		for key, value := range *urlOptions.Query {
			queryParams += fmt.Sprintf("%s=%s&", key, value)
		}
		queryParams = strings.TrimSuffix(queryParams, "&")
	}

	if urlOptions != nil && urlOptions.Subdomain != "" {
		baseUrl = fmt.Sprintf("https://%s.fal.run/%s/", urlOptions.Subdomain, urlOptions.AppId)
	}

	return baseUrl + route + queryParams
}

func (r *Client) newRequest(ctx context.Context, method, path string, body io.Reader, urlOptions *UrlOptions) (*http.Request, error) {
	url := constructUrl(r.options.baseUrl, path, urlOptions)
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

// Fetch makes an HTTP request to FAL's API.
func (r *Client) Fetch(ctx context.Context, method, path string, body interface{}, out interface{}, urlOptions *UrlOptions) error {
	bodyBuffer := &bytes.Buffer{}
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyBuffer = bytes.NewBuffer(bodyBytes)
	}
	req, err := r.newRequest(ctx, method, path, bodyBuffer, urlOptions)
	if err != nil {
		return err
	}

	return r.do(req, out)
}

// shouldRetry returns true if the request should be retried.
//
// - GET requests should be retried if the response status code is 429 or 5xx.
// - Other requests should be retried if the response status code is 429.
func (r *Client) shouldRetry(response *http.Response, method string) bool {
	if method == http.MethodGet {
		return response.StatusCode == 429 || (response.StatusCode >= 500 && response.StatusCode < 600)
	}

	return response.StatusCode == 429
}

// do makes an HTTP request to FAL's API.
func (r *Client) do(request *http.Request, out interface{}) error {
	maxRetries := r.options.retryPolicy.maxRetries
	backoff := r.options.retryPolicy.backoff

	attempts := 0
	var apiError *APIError

	for ok := true; ok; ok = attempts < maxRetries {
		resp, err := r.c.Do(request)
		if err != nil || resp == nil {
			return fmt.Errorf("failed to make request: %w", err)
		}

		defer resp.Body.Close()
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			apiError = unmarshalAPIError(resp, responseBytes)

			if !r.shouldRetry(resp, request.Method) {
				return apiError
			}

			delay := backoff.NextDelay(attempts)

			if delay > 0 {
				time.Sleep(delay)
			}

			attempts++
		} else {
			if out != nil {
				if err := json.Unmarshal(responseBytes, &out); err != nil {
					return fmt.Errorf("failed to unmarshal response: %w", err)
				}
			}

			return nil
		}
	}

	if attempts > 0 {
		return fmt.Errorf("request failed after %d attempts", maxRetries)
	}

	return fmt.Errorf("request failed")
}
