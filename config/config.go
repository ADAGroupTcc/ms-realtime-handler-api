package config

import (
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Environments define the environment variables
type Environments struct {
	APIPort                            string `envconfig:"API_PORT"`
	AppEnv                             string `envconfig:"APP_ENV"`
	AppName                            string `envconfig:"APP_NAME"`
	LoggerLevel                        string `envconfig:"LOGGER_LEVEL"`
	InstrumentationService             string `envconfig:"INSTRUMENTATION_SERVICE"`
	InstrumentationTracesEndpoint      string `envconfig:"INSTRUMENTATION_TRACES_ENDPOINT"`
	InstrumentationMetricsEndpoint     string `envconfig:"INSTRUMENTATION_METRICS_ENDPOINT"`
	InstrumentationSendMetricsInterval int    `envconfig:"INSTRUMENTATION_SEND_METRICS_INTERVAL"`
	RedisHost                          string `envconfig:"REDIS_HOST"`
	RedisPoolSize                      int    `envconfig:"REDIS_POOL_SIZE"`
}

// LoadEnvVars load the environment variables
func LoadEnvVars(log *logger.Logger) *Environments {
	godotenv.Load()
	c := &Environments{}
	if err := envconfig.Process("", c); err != nil {
		log.Fatal("Failed to load environment variables", err)
		return nil
	}
	return c
}
