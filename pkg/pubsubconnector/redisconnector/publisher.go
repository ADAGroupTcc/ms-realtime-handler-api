package redisconnector

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/pubsubconnector"
	redis "github.com/go-redis/redis/v8"
)

type redisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(config *redisConnectionConfig) pubsubconnector.Publisher {
	return &redisPublisher{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			PoolSize: config.PoolSize, // Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
		}),
	}
}

type publisherConfig struct {
	topic   string
	key     string
	headers map[string]string
}

func getConfigs(configMap *map[string]interface{}) (publisherConfig, error) {
	config := publisherConfig{}
	if configMap == nil {
		return config, errors.New("redis config must not be null")
	}

	topic, ok := (*configMap)["topic"].(string)
	if !ok {
		return config, errors.New("redis config.topic must be a string")
	}

	if topic == "" {
		return config, errors.New("redis config.topic must not be empty")
	}

	config.topic = topic

	return config, nil
}

func (redisPublisher *redisPublisher) Publish(ctx context.Context, message interface{}, configMap *map[string]interface{}) error {
	data, _ := json.Marshal(message)

	config, err := getConfigs(configMap)
	if err != nil {
		return err
	}

	resp := redisPublisher.client.Publish(ctx, config.topic, data)
	if resp.Err() != nil {
		return resp.Err()
	}
	return nil
}
