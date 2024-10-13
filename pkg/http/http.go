package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

const (
	defaultTimeout         = 2 * time.Second
	defaultRetryAfter      = 1 * time.Second
	defaultMaxIdleConns    = 100
	defaultMaxConnsPerHost = 100
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
type HttpClient struct {
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

	return &HttpClient{
		client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:    config.MaxIdleConns,
				MaxConnsPerHost: config.MaxConnsPerHost,
			},
		},
		config: &config,
	}, nil
}

// Get performs an HTTP GET request.
func (c *HttpClient) Get(ctx context.Context, clientConfig ClientConfig) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodGet, clientConfig, nil)
}

// Post performs an HTTP POST request.
func (c *HttpClient) Post(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodPost, clientConfig, payload)
}

// Patch performs an HTTP PATCH request.
func (c *HttpClient) Patch(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodPatch, clientConfig, payload)
}

// Put performs an HTTP PUT request.
func (c *HttpClient) Put(ctx context.Context, clientConfig ClientConfig, payload []byte) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodPut, clientConfig, payload)
}

// Delete performs an HTTP DELETE request.
func (c *HttpClient) Delete(ctx context.Context, clientConfig ClientConfig) (*HttpResponse, error) {
	return c.execute(ctx, http.MethodDelete, clientConfig, nil)
}

func (c *HttpClient) execute(ctx context.Context, method string, clientConfig ClientConfig, payload []byte) (*HttpResponse, error) {
	url := c.formatUrl(clientConfig.Endpoint)

	var body io.Reader
	if payload != nil {
		body = bytes.NewBuffer(payload)
	} else {
		body = nil
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return &HttpResponse{}, err
	}

	response, err := c.executeRequest(ctx, req, clientConfig)

	if err != nil {
		return &HttpResponse{}, err
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return &HttpResponse{}, err
	}

	headers := response.Header.Clone()

	return &HttpResponse{
		Body:       responseBody,
		Headers:    headers,
		StatusCode: response.StatusCode,
	}, nil
}

func (c *HttpClient) formatUrl(endpoint string) string {
	return fmt.Sprintf("%s%s", c.config.BaseURL, endpoint)
}

func (c *HttpClient) setHeaders(ctx context.Context, req *http.Request, customHeaders map[string]string) {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Request-Id", uuid.New().String())

	if requestId, ok := ctx.Value("requestid").(string); ok {
		req.Header.Set("X-Request-Id", requestId)
	}

	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}
}

func (c *HttpClient) executeRequest(ctx context.Context, req *http.Request, clientConfig ClientConfig) (res *http.Response, err error) {
	var response *http.Response

	c.setHeaders(ctx, req, clientConfig.Headers)
	// req = c.config.Instrument.RequestWithTransactionContext(ctx, req)
	response, err = c.executeRetryableRequest(ctx, req, clientConfig)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *HttpClient) executeRetryableRequest(ctx context.Context, req *http.Request, clientConfig ClientConfig) (*http.Response, error) {
	var response *http.Response
	var err error

	retryableRequest := req

	for i := 1; i <= c.config.Retries; i++ {
		if retryableRequest.Method != http.MethodGet {
			retryableRequest.Body, _ = req.GetBody()
		}
		response, err = c.client.Do(retryableRequest)

		if err != nil {
			if os.IsTimeout(err) {
				if shouldRetry(c.config.RetryWhenStatus, 408) {
					retryTime := exponentialBackoff(i, c.config.RetryAfter)
					time.Sleep(retryTime)
					continue
				}
			}
			return nil, err
		}

		if shouldRetry(c.config.RetryWhenStatus, response.StatusCode) {
			retryTime := exponentialBackoff(i, c.config.RetryAfter)
			time.Sleep(retryTime)
			continue
		}

		return response, nil
	}

	return response, err
}
