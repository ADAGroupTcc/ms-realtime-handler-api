package http

import (
	"errors"
	"math/rand"
	"time"
)

type RetryConfig struct {
	Retries         int
	RetryAfter      time.Duration
	RetryWhenStatus []int
}

type ClientConfig struct {
	MetricUrl string
	Endpoint  string
	Headers   map[string]string
}

type Config struct {
	BaseURL         string
	Timeout         time.Duration
	MaxIdleConns    int
	MaxConnsPerHost int
	RetryConfig
}

func (c *Config) validateConfig() error {
	if c.MaxConnsPerHost < 0 {
		return errors.New("http_config: maxConnsPerHost could not be less than zero")
	}

	if c.MaxIdleConns < 0 {
		return errors.New("http_config: maxIdleConns could not be less than zero")
	}

	if c.Timeout.Seconds() < 0 {
		return errors.New("http_config: Timeout could not be negative")
	}

	if c.Retries < 0 {
		return errors.New("http_config: Retries could not be negative")
	}

	if c.RetryAfter.Seconds() < 0 {
		return errors.New("http_config: RetryAfter could not be negative")
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

	if c.Timeout.Seconds() == 0 {
		c.Timeout = defaultTimeout
	}

	if c.RetryAfter.Seconds() == 0 {
		c.RetryAfter = defaultRetryAfter
	}
}

func shouldRetry(rules []int, statusCode int) bool {
	for _, code := range rules {
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
	const jitterFactor = 0.2 // 20% jitter
	jitter := time.Duration(float64(delay) * jitterFactor * (rand.Float64()*2 - 1))
	return delay + jitter
}
