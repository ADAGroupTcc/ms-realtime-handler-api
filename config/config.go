package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"

	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/spf13/viper"
)

// Environments define the environment variables
type Environments struct {
	APIPort                            string `mapstructure:"API_PORT"`
	AppEnv                             string `mapstructure:"APP_ENV"`
	AppName                            string `mapstructure:"APP_NAME"`
	LoggerLevel                        string `mapstructure:"LOGGER_LEVEL"`
	InstrumentationService             string `mapstructure:"INSTRUMENTATION_SERVICE"`
	InstrumentationTracesEndpoint      string `mapstructure:"INSTRUMENTATION_TRACES_ENDPOINT"`
	InstrumentationMetricsEndpoint     string `mapstructure:"INSTRUMENTATION_METRICS_ENDPOINT"`
	InstrumentationSendMetricsInterval string `mapstructure:"INSTRUMENTATION_SEND_METRICS_INTERVAL"`
}

// LoadEnvVars load the environment variables
func LoadEnvVars(l *logger.Logger) *Environments {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetDefault("API_PORT", "8080")
	env := &Environments{}

	if err := viper.ReadInConfig(); err != nil {
		l.Warnf("unable find or read configuration file: %w", err)
	}
	if err := loadMappedEnvVariables(env); err != nil {
		l.Warnf("unable to load environment variables: %w", err)
	}
	if err := viper.Unmarshal(&env); err != nil {
		l.Warnf("unable marshal configuration file: %w", err)
	}
	setEnvVars(viper.AllSettings())

	return env
}

func setEnvVars(s map[string]any) {
	for k, v := range s {
		if os.Getenv(strings.ToUpper(k)) == "" {
			os.Setenv(strings.ToUpper(k), fmt.Sprintf("%v", v))
		}
	}
}

func loadMappedEnvVariables(env *Environments) error {
	envKeysMap := &map[string]any{}
	if err := mapstructure.Decode(env, &envKeysMap); err != nil {
		return err
	}
	for k := range *envKeysMap {
		if err := viper.BindEnv(k); err != nil {
			return err
		}
	}
	return nil
}
