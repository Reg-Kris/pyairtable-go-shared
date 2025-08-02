// Package config provides configuration management utilities using Viper
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host         string `mapstructure:"host" default:"localhost"`
	Port         int    `mapstructure:"port" default:"8080"`
	ReadTimeout  int    `mapstructure:"read_timeout" default:"30"`
	WriteTimeout int    `mapstructure:"write_timeout" default:"30"`
	Environment  string `mapstructure:"environment" default:"development"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host         string `mapstructure:"host" default:"localhost"`
	Port         int    `mapstructure:"port" default:"5432"`
	User         string `mapstructure:"user" default:"postgres"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database" default:"pyairtable"`
	SSLMode      string `mapstructure:"ssl_mode" default:"disable"`
	MaxOpenConns int    `mapstructure:"max_open_conns" default:"25"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" default:"25"`
	MaxLifetime  int    `mapstructure:"max_lifetime" default:"300"`
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Host         string `mapstructure:"host" default:"localhost"`
	Port         int    `mapstructure:"port" default:"6379"`
	Password     string `mapstructure:"password"`
	Database     int    `mapstructure:"database" default:"0"`
	PoolSize     int    `mapstructure:"pool_size" default:"10"`
	MinIdleConns int    `mapstructure:"min_idle_conns" default:"5"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"`
	JWTExpiration int    `mapstructure:"jwt_expiration" default:"3600"`
	Issuer        string `mapstructure:"issuer" default:"pyairtable"`
}

// LoggerConfig contains logger configuration
type LoggerConfig struct {
	Level      string `mapstructure:"level" default:"info"`
	Format     string `mapstructure:"format" default:"json"`
	OutputPath string `mapstructure:"output_path" default:"stdout"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" default:"true"`
	Port    int    `mapstructure:"port" default:"9090"`
	Path    string `mapstructure:"path" default:"/metrics"`
}

// Load loads configuration from various sources
func Load(configPath string) (*Config, error) {
	v := viper.New()
	
	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/pyairtable")
	}
	
	// Set environment variable prefix
	v.SetEnvPrefix("PYAIRTABLE")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Set defaults
	setDefaults(v)
	
	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.environment", "development")
	
	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.database", "pyairtable")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.max_lifetime", 300)
	
	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	
	// Auth defaults
	v.SetDefault("auth.jwt_expiration", 3600)
	v.SetDefault("auth.issuer", "pyairtable")
	
	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.output_path", "stdout")
	
	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9090)
	v.SetDefault("metrics.path", "/metrics")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}
	
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	
	return nil
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Server.Environment) == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Server.Environment) == "production"
}