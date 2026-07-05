package config

import (
	"errors"
	"fmt"
	"reflect"

	casparcg "github.com/overlayfox/caspaw-cg/src/caspar"
	"github.com/overlayfox/caspaw-cg/src/data"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Config struct {
	DataSourceManager *data.Config     `mapstructure:"data_source_manager"`
	CasparCGClient    *casparcg.Config `mapstructure:"casparcg_client"`
}

type Defaulter interface {
	Default()
}

type Validator interface {
	Validate() error
}

var defaulterType = reflect.TypeOf((*Defaulter)(nil)).Elem()

// applyDefaults checks the main struct and its immediate fields for the Defaulter interface
func applyDefaults(target any) {
	if d, ok := target.(Defaulter); ok {
		d.Default()
	}
	v := reflect.ValueOf(target).Elem()
	if v.Kind() == reflect.Struct {
		for i := range v.NumField() {
			field := v.Field(i)
			if !field.CanSet() {
				continue
			}
			if field.Kind() == reflect.Pointer {
				if field.IsNil() {
					if !field.Type().Implements(defaulterType) {
						continue
					}
					field.Set(reflect.New(field.Type().Elem()))
				}
				if d, ok := field.Interface().(Defaulter); ok {
					d.Default()
				}
				continue
			}
			if field.CanAddr() {
				if d, ok := field.Addr().Interface().(Defaulter); ok {
					d.Default()
				}
			}
		}
	}
}

// applyValidation checks the main struct and its immediate fields for the Validator interface
func applyValidation(target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.CanAddr() {
		if validator, ok := v.Addr().Interface().(Validator); ok {
			if err := validator.Validate(); err != nil {
				return err
			}
		}
	} else if validator, ok := v.Interface().(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Field(i)
			if err := applyValidation(field.Addr().Interface()); err != nil {
				fieldName := v.Type().Field(i).Name
				return fmt.Errorf("[%s] %w", fieldName, err)
			}
		}

	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			element := v.Index(i)
			if element.Kind() == reflect.Struct {
				if err := applyValidation(element.Addr().Interface()); err != nil {
					return fmt.Errorf("index %d: %w", i, err)
				}
			}
		}
	}

	return nil
}

func LoadConfig(logger zerolog.Logger) (*Config, error) {
	var cfg Config
	applyDefaults(&cfg)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

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
