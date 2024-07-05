package fal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
)

type mockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestNewClient(t *testing.T) {
	t.Run("WithToken", func(t *testing.T) {
		client, err := NewClient(WithToken("test-token"))
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if client.options.auth != "test-token" {
			t.Errorf("Expected auth token to be 'test-token', got %s", client.options.auth)
		}
	})

	t.Run("WithTokenFromEnv", func(t *testing.T) {
		os.Setenv(envAuthToken, "env-token")
		defer os.Unsetenv(envAuthToken)

		client, err := NewClient(WithTokenFromEnv())
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if client.options.auth != "env-token" {
			t.Errorf("Expected auth token to be 'env-token', got %s", client.options.auth)
		}
	})

	t.Run("WithUserAgent", func(t *testing.T) {
		client, err := NewClient(WithToken("test-token"), WithUserAgent("custom-agent"))
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if client.options.userAgent != "custom-agent" {
			t.Errorf("Expected user agent to be 'custom-agent', got %s", client.options.userAgent)
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		_, err := NewClient()
		if err != ErrNoAuth {
			t.Fatalf("Expected ErrNoAuth, got %v", err)
		}
	})
}

func TestConstructUrl(t *testing.T) {
	baseUrl := "https://api.example.com"
	route := "/test"

	t.Run("BasicUrl", func(t *testing.T) {
		url := constructUrl(baseUrl, route, nil)
		expected := "https://api.example.com/test"
		if url != expected {
			t.Errorf("Expected URL %s, got %s", expected, url)
		}
	})

	t.Run("WithSubdomain", func(t *testing.T) {
		options := &UrlOptions{
			Subdomain: "sub",
			AppId:     "app123",
		}
		url := constructUrl(baseUrl, route, options)
		expected := "https://sub.fal.run/app123/test"
		if url != expected {
			t.Errorf("Expected URL %s, got %s", expected, url)
		}
	})

	t.Run("WithQueryParams", func(t *testing.T) {
		query := QueryParams{"param1": "value1", "param2": "value2"}
		options := &UrlOptions{Query: &query}
		url := constructUrl(baseUrl, route, options)
		expected := "https://api.example.com/test?param1=value1&param2=value2"
		if url != expected {
			t.Errorf("Expected URL %s, got %s", expected, url)
		}
	})
}

func TestClientFetch(t *testing.T) {
	mockResponse := map[string]string{"message": "success"}
	mockResponseBody, _ := json.Marshal(mockResponse)

	mockTripper := &mockRoundTripper{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBuffer(mockResponseBody)),
			}, nil
		},
	}

	mockHttpClient := &http.Client{Transport: mockTripper}

	client, _ := NewClient(
		WithToken("test-token"),
		WithUserAgent("test-agent"),
		WithHttpClient(mockHttpClient),
	)

	var response map[string]string
	err := client.Fetch(context.Background(), "GET", "/test", nil, &response, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(response, mockResponse) {
		t.Errorf("Expected response %v, got %v", mockResponse, response)
	}
}

func TestClientFetchWithBody(t *testing.T) {
	requestBody := map[string]string{"key": "value"}
	mockResponse := map[string]string{"message": "received"}
	mockResponseBody, _ := json.Marshal(mockResponse)

	mockTripper := &mockRoundTripper{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			var receivedBody map[string]string
			json.NewDecoder(req.Body).Decode(&receivedBody)
			if !reflect.DeepEqual(receivedBody, requestBody) {
				t.Errorf("Expected request body %v, got %v", requestBody, receivedBody)
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBuffer(mockResponseBody)),
			}, nil
		},
	}

	mockHttpClient := &http.Client{Transport: mockTripper}

	client, _ := NewClient(
		WithToken("test-token"),
		WithUserAgent("test-agent"),
		WithHttpClient(mockHttpClient),
	)

	var response map[string]string
	err := client.Fetch(context.Background(), "POST", "/test", requestBody, &response, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(response, mockResponse) {
		t.Errorf("Expected response %v, got %v", mockResponse, response)
	}
}
