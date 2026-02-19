package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Docker   DockerConfig   `mapstructure:"docker"`
	Keycloak KeycloakConfig `mapstructure:"keycloak"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type LLMConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type DockerConfig struct {
	Host string `mapstructure:"host"`
}

type KeycloakConfig struct {
	URL     string `mapstructure:"url"`
	JWKSURL string `mapstructure:"jwks_url"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

func Load() (*Config, error) {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("docker.host", "unix:///var/run/docker.sock")
	viper.SetDefault("redis.url", "redis://localhost:6379")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
