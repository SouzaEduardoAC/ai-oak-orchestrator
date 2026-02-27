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
	Valkey   ValkeyConfig   `mapstructure:"valkey"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
}

type DockerConfig struct {
	Host string `mapstructure:"host"`
}

type KeycloakConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
	JWKSURL string `mapstructure:"jwks_url"`
}

type ValkeyConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
}

func Load() (*Config, error) {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("docker.host", "unix:///var/run/docker.sock")
	viper.SetDefault("valkey.url", "valkey://localhost:6379")
	viper.SetDefault("keycloak.enabled", "true")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read .env file if it exists
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	// Explicit bindings and manual mapping for .env compatibility
	mapEnv := func(key, env string) {
		viper.BindEnv(key, env)
		if val := viper.GetString(env); val != "" {
			viper.Set(key, val)
		}
	}

	mapEnv("valkey.url", "VALKEY_URL")
	mapEnv("valkey.password", "VALKEY_PASSWORD")
	mapEnv("server.port", "SERVER_PORT")
	mapEnv("llm.provider", "LLM_PROVIDER")
	mapEnv("llm.api_key", "LLM_API_KEY")
	mapEnv("llm.model", "LLM_MODEL")
	mapEnv("keycloak.enabled", "KEYCLOAK_ENABLED")
	mapEnv("keycloak.url", "KEYCLOAK_URL")
	mapEnv("keycloak.jwks_url", "KEYCLOAK_JWKS_URL")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
