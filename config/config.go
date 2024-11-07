package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Environments define the environment variables
type Environments struct {
	APIPort string `envconfig:"PORT"`
	AppName string `envconfig:"APP_NAME"`

	RedisHost                  string `envconfig:"REDIS_HOST"`
	RedisPoolSize              int    `envconfig:"REDIS_POOL_SIZE"`
	RedisSubscribeTopic        string `envconfig:"REDIS_SUBSCRIBER_TOPIC"`
	WsReadDeadlineAwaitSeconds int    `envconfig:"WS_READ_DEADLINE_AWAIT_SECONDS" default:"10"`

	MessagesApiUrl string `envconfig:"MESSAGES_API_URL"`

	SorterApiUrl string `envconfig:"SORTER_API_URL"`
}

// LoadEnvVars load the environment variables
func LoadEnvVars() *Environments {
	godotenv.Load()
	c := &Environments{}
	if err := envconfig.Process("", c); err != nil {
		return nil
	}
	return c
}
