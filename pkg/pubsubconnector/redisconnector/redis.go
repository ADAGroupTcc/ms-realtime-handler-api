package redisconnector

import "errors"

type redisConnectionConfig struct {
	Addr     string
	PoolSize int
}

func (redisConfig *redisConnectionConfig) ValidateConfig() error {
	if redisConfig.Addr == "" {
		return errors.New("redis_config: address is required")
	}

	if redisConfig.PoolSize == 0 {
		redisConfig.PoolSize = 10
	}

	return nil
}

func NewConfig(addr string, poolSize int) *redisConnectionConfig {
	return &redisConnectionConfig{
		Addr:     addr,
		PoolSize: poolSize,
	}
}
