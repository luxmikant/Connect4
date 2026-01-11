package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Kafka       KafkaConfig    `mapstructure:"kafka"`
	Redis       RedisConfig    `mapstructure:"redis"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int      `mapstructure:"port"`
	Host         string   `mapstructure:"host"`
	CORSOrigins  []string `mapstructure:"cors_origins"`
	ReadTimeout  int      `mapstructure:"read_timeout"`
	WriteTimeout int      `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string `mapstructure:"url"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	SSLMode         string `mapstructure:"ssl_mode"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	BootstrapServers string `mapstructure:"bootstrap_servers"`
	APIKey           string `mapstructure:"api_key"`
	APISecret        string `mapstructure:"api_secret"`
	Topic            string `mapstructure:"topic"`
	ConsumerGroup    string `mapstructure:"consumer_group"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	// Load .env file first (if it exists)
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't fail if it doesn't exist
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: Could not load .env file: %v\n", err)
		}
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set default values
	setDefaults()

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind specific environment variables to config keys
	_ = viper.BindEnv("database.url", "DATABASE_URL")
	_ = viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	_ = viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")

	viper.BindEnv("kafka.bootstrap_servers", "KAFKA_BOOTSTRAP_SERVERS")
	viper.BindEnv("kafka.api_key", "KAFKA_API_KEY")
	viper.BindEnv("kafka.api_secret", "KAFKA_API_SECRET")
	viper.BindEnv("kafka.topic", "KAFKA_TOPIC")
	viper.BindEnv("kafka.consumer_group", "KAFKA_CONSUMER_GROUP")

	viper.BindEnv("redis.url", "REDIS_URL")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.cors_origins", "SERVER_CORS_ORIGINS")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	viper.BindEnv("environment", "ENVIRONMENT")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Environment
	viper.SetDefault("environment", "development")

	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.cors_origins", []string{"*"})
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Database defaults
	viper.SetDefault("database.url", "postgres://postgres:password@localhost:5432/connect4?sslmode=disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", 300)
	viper.SetDefault("database.ssl_mode", "disable")

	// Kafka defaults
	viper.SetDefault("kafka.bootstrap_servers", "localhost:9092")
	viper.SetDefault("kafka.topic", "game-events")
	viper.SetDefault("kafka.consumer_group", "analytics-service")

	// Redis defaults
	viper.SetDefault("redis.url", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	if config.Kafka.BootstrapServers == "" {
		return fmt.Errorf("kafka bootstrap servers are required")
	}

	return nil
}
