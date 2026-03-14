package config

import (
	"errors"
	"fmt"
	"reflect"

	"caspaw-cg/src/data"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Defaulter interface {
	Default()
}

type Validator interface {
	Validate() error
}

// applyDefaults checks the main struct and its immediate fields for the Defaulter interface
func applyDefaults(target any) {
	if d, ok := target.(Defaulter); ok {
		d.Default()
	}

	v := reflect.ValueOf(target).Elem()
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanAddr() {
				if d, ok := field.Addr().Interface().(Defaulter); ok {
					d.Default()
				}
			}
		}
	}
}

// applyValidation checks the main struct and its immediate fields for the Validator interface
func applyValidation(target interface{}) error {
	if v, ok := target.(Validator); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	val := reflect.ValueOf(target).Elem()
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if field.CanAddr() {
				if v, ok := field.Addr().Interface().(Validator); ok {
					if err := v.Validate(); err != nil {
						fieldName := val.Type().Field(i).Name
						return fmt.Errorf("'%s' validation failed: %w", fieldName, err)
					}
				}
			}
		}
	}
	return nil
}

type Config struct {
	DataSourceManager data.Config `yaml:"data_source_manager"`
}

func LoadConfig(logger zerolog.Logger) (*Config, error) {
	var cfg Config
	applyDefaults(&cfg)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		logger.Info().Msg("no config file found, relying on defaults")
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := applyValidation(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
