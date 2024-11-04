package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type HttpClienter interface {
	Get(ctx context.Context, clientConfig ClientConfig) (*HttpResponse, error)
	Post(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error)
	Patch(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error)
	Put(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error)
	Delete(ctx context.Context, clientConfig ClientConfig) (*HttpResponse, error)
}

// HttpResponse represents the return of the HttpClient.
type HttpResponse struct {
	Body       []byte
	Headers    map[string][]string
	StatusCode int
}

// HttpClient represents a custom HTTP client.
type httpClient struct {
	client *http.Client
	config *Config
}

// New creates a new custom HTTP client with the given configuration.
func New(config Config) (HttpClienter, error) {
	err := config.validateConfig()
	if err != nil {
		return nil, err
	}

	config.normalizeConfig()

	return &httpClient{
		client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:    config.MaxIdleConns,
				MaxConnsPerHost: config.MaxConnsPerHost,
				IdleConnTimeout: config.IdleConnTimeout,
			},
		},
		config: &config,
	}, nil
}

// Get performs an HTTP GET request.
func (c *httpClient) Get(ctx context.Context, clientConfig ClientConfig) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodGet, clientConfig, nil)
}

// Post performs an HTTP POST request.
func (c *httpClient) Post(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodPost, clientConfig, payload)
}

// Patch performs an HTTP PATCH request.
func (c *httpClient) Patch(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodPatch, clientConfig, payload)
}

// Put performs an HTTP PUT request.
func (c *httpClient) Put(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodPut, clientConfig, payload)
}

// Delete performs an HTTP DELETE request.
func (c *httpClient) Delete(ctx context.Context, clientConfig ClientConfig) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodDelete, clientConfig, nil)
}

func (c *httpClient) execute(ctx context.Context, method string, clientConfig ClientConfig, payload []byte) (*HttpResponse, error) {
	url := c.formatUrl(clientConfig.Endpoint)

	var body io.Reader
	if payload != nil {
		body = bytes.NewBuffer(payload)
	} else {
		body = nil
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	response, err := c.executeRequest(ctx, req, clientConfig)

	if err != nil {
		return nil, err
	}

	if response.Body == nil {
		err := fmt.Errorf("response body is nil")
		return nil, err
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	headers := response.Header.Clone()

	return &HttpResponse{
		Body:       responseBody,
		Headers:    headers,
		StatusCode: response.StatusCode,
	}, nil
}

func (c *httpClient) formatUrl(endpoint string) string {
	return fmt.Sprintf("%s%s", c.config.BaseURL, endpoint)
}

func (c *httpClient) setHeaders(ctx context.Context, req *http.Request, customHeaders map[string]string) {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Request-Id", uuid.New().String())

	if requestId, ok := ctx.Value("X-Request-Id").(string); ok {
		req.Header.Set("X-Request-Id", requestId)
	}

	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}
}

func (c *httpClient) executeRequest(ctx context.Context, req *http.Request, clientConfig ClientConfig) (res *http.Response, err error) {
	c.setHeaders(ctx, req, clientConfig.Headers)
	response, err := c.executeRetryableRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *httpClient) executeRetryableRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	retryableRequest := req

	var retryTime time.Duration
	var err error
	var response *http.Response
	retriesAttempts := c.config.Retries
	reason := ""

	for attempts := 0; attempts <= retriesAttempts; attempts++ {
		if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
			if req.Body != nil {
				var getBodyErr error
				retryableRequest.Body, getBodyErr = req.GetBody()
				if getBodyErr != nil {
					return nil, getBodyErr
				}
			}
		}

		if retryableRequest.Method != http.MethodGet && retryableRequest.Method != http.MethodDelete {
			retryableRequest.Body, _ = req.GetBody()
		}

		if attempts > 0 {
		}

		response, err = c.client.Do(retryableRequest)

		if err != nil {
			reason = err.Error()
			if isTimeout(err) {
				if c.config.shouldRetry(http.StatusRequestTimeout) {
					retryTime = c.defineRetryInterval(attempts)
					if retriesAttempts > 0 {
						time.Sleep(retryTime)
					}
					continue
				}
			}
			return nil, err
		}

		reason = fmt.Sprintf("status code %d", response.StatusCode)
		if c.config.shouldRetry(response.StatusCode) {
			retryTime = c.defineRetryInterval(attempts)
			if retriesAttempts > 0 {
				time.Sleep(retryTime)
			}
			continue
		}

		return response, nil
	}

	if retriesAttempts == 0 {
		return nil, errors.New("error on execute request: " + reason)
	}

	return nil, errors.New("retries attempts exhausted")
}

func isTimeout(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return false
}

func (c *httpClient) defineRetryInterval(attempts int) time.Duration {
	if c.config.ExponentialBackoffEnabled {
		return exponentialBackoff(attempts, c.config.RetryAfter)
	}
	return c.config.RetryAfter
}
