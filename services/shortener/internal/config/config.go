package config

import (
	"github.com/spf13/viper"
)

// Config holds all configuration for the service
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	MongoDB  MongoDBConfig  `mapstructure:"mongodb"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Features FeatureConfig  `mapstructure:"features"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            string `mapstructure:"port"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// MongoDBConfig holds MongoDB connection configuration
type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	URI      string `mapstructure:"uri"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// FeatureConfig holds feature flag configuration
type FeatureConfig struct {
	UnleashURL string `mapstructure:"unleash_url"`
	APIToken   string `mapstructure:"api_token"`
}

// Load reads configuration from environment variables and files
func Load() (*Config, error) {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.shutdown_timeout", 30)
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "urlshortener")
	viper.SetDefault("redis.uri", "localhost:6379")
	viper.SetDefault("redis.db", 0)

	viper.SetEnvPrefix("URL_SHORTENER")
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
