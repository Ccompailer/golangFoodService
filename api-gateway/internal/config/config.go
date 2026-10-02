package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Services ServicesConfig `mapstructure:"services"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Environment  string `mapstructure:"environment"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type AuthConfig struct {
	JWKSURL    string        `mapstructure:"jwks_url"`
	Issuer     string        `mapstructure:"issuer"`
	Audience   string        `mapstructure:"audience"`
	Skew       time.Duration `mapstructure:"skew"`
	RefreshMin time.Duration `mapstructure:"refresh_min"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ServicesConfig struct {
	UserService    string `mapstructure:"user_service"`
	OrderService   string `mapstructure:"order_service"`
	CatalogService string `mapstructure:"catalog_service"`
	PaymentService string `mapstructure:"payment_service"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", 10)
	viper.SetDefault("server.write_timeout", 10)
	viper.SetDefault("auth.access_token_ttl", 3600)
	viper.SetDefault("auth.refresh_token_ttl", 604800)
	viper.SetDefault("log.level", "info")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
