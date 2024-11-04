package http

import (
	"errors"
	"math/rand"
	"time"
)

const (
	defaultTimeout         = 10000 * time.Millisecond
	defaultRetryAfter      = 600 * time.Millisecond
	defaultMaxIdleConns    = 200
	defaultMaxConnsPerHost = 250
	defaultIdleConnTimeout = 90 * time.Second
)

type RetryConfig struct {
	Retries                   int
	RetryAfter                time.Duration
	RetryWhenStatus           []int
	ExponentialBackoffEnabled bool
}

type ClientConfig struct {
	MetricUrl string
	Endpoint  string
	Headers   map[string]string
}

type Config struct {
	BaseURL           string
	AllowEmptyBaseUrl bool
	Timeout           time.Duration
	IdleConnTimeout   time.Duration
	MaxIdleConns      int
	MaxConnsPerHost   int
	RetryConfig
}

func (c *Config) validateConfig() error {
	if c.MaxConnsPerHost < 0 {
		return errors.New("http_config: maxConnsPerHost could not be less than zero")
	}

	if c.MaxIdleConns < 0 {
		return errors.New("http_config: maxIdleConns could not be less than zero")
	}

	if c.IdleConnTimeout < 0 {
		return errors.New("http_config: idleConnTimeout could not be less than zero")
	}

	if c.Timeout.Milliseconds() < 0 {
		return errors.New("http_config: Timeout could not be negative")
	}

	if c.Retries < 0 {
		return errors.New("http_config: Retries could not be negative")
	}

	if c.RetryAfter.Milliseconds() < 0 {
		return errors.New("http_config: RetryAfter could not be negative")
	}

	if !c.AllowEmptyBaseUrl && c.BaseURL == "" {
		return errors.New("http_config: BaseURL could not be empty")
	}

	return nil
}

func (c *Config) normalizeConfig() {
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = defaultMaxIdleConns
	}

	if c.MaxConnsPerHost == 0 {
		c.MaxConnsPerHost = defaultMaxConnsPerHost
	}

	if c.IdleConnTimeout.Milliseconds() == 0 {
		c.IdleConnTimeout = defaultIdleConnTimeout
	}

	if c.Timeout.Milliseconds() == 0 {
		c.Timeout = defaultTimeout
	}

	if c.RetryAfter.Milliseconds() == 0 {
		c.RetryAfter = defaultRetryAfter
	}
}

func (c *Config) shouldRetry(statusCode int) bool {
	for _, code := range c.RetryWhenStatus {
		if code == statusCode {
			return true
		}
	}
	return false
}

func exponentialBackoff(attempt int, baseDelay time.Duration) time.Duration {
	delay := baseDelay * time.Duration(1<<uint(attempt))
	return jitter(delay)
}

func jitter(delay time.Duration) time.Duration {
	const jitterFactor = 0.1 // 10% jitter
	jitter := time.Duration(float64(delay) * jitterFactor * (rand.Float64()*2 - 1))
	return delay + jitter
}
