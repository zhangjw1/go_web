package config

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Kafka      KafkaConfig      `mapstructure:"kafka"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	Blockchain BlockchainConfig `mapstructure:"blockchain"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	LogLevel        string `mapstructure:"log_level"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Brokers []string `mapstructure:"brokers"`
	GroupID string   `mapstructure:"group_id"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// BlockchainConfig holds blockchain configuration
type BlockchainConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	NetworkURL  string `mapstructure:"network_url"`
	NetworkName string `mapstructure:"network_name"`
	Timeout     int    `mapstructure:"timeout"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := &Config{}

	// Set default values
	setDefaults()

	// Set config file path
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults and env vars
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			// Check if it's a file not found error when specific path is provided
			if configPath != "" && strings.Contains(err.Error(), "no such file or directory") || strings.Contains(err.Error(), "cannot find the file") {
				fmt.Println("Config file not found, using defaults and environment variables")
			} else {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Database defaults
	viper.SetDefault("database.enabled", true)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "go_web_starter")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.log_level", "error")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", 60)

	// Redis defaults
	viper.SetDefault("redis.enabled", true)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.database", 0)

	// Kafka defaults
	viper.SetDefault("kafka.enabled", true)
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.group_id", "go-web-starter")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.filename", "logs/app.log")
	viper.SetDefault("logger.max_size", 100)
	viper.SetDefault("logger.max_backups", 3)
	viper.SetDefault("logger.max_age", 28)
	viper.SetDefault("logger.compress", true)

	// Blockchain defaults
	viper.SetDefault("blockchain.enabled", true)
	viper.SetDefault("blockchain.network_url", "https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
	viper.SetDefault("blockchain.network_name", "ethereum")
	viper.SetDefault("blockchain.timeout", 30)
}

// validate validates the configuration
func validate(config *Config) error {
	// Validate server config
	if config.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if config.Server.Mode != "debug" && config.Server.Mode != "release" && config.Server.Mode != "test" {
		return fmt.Errorf("server mode must be one of: debug, release, test")
	}

	// Validate database config (only when enabled)
	if config.Database.Enabled {
		if config.Database.Host == "" {
			return fmt.Errorf("database host is required")
		}
		if config.Database.Username == "" {
			return fmt.Errorf("database username is required")
		}
		if config.Database.Database == "" {
			return fmt.Errorf("database name is required")
		}
	}

	// Validate Redis config (only when enabled)
	if config.Redis.Enabled {
		if config.Redis.Host == "" {
			return fmt.Errorf("redis host is required")
		}
	}

	// Validate Kafka config (only when enabled)
	if config.Kafka.Enabled {
		if len(config.Kafka.Brokers) == 0 {
			return fmt.Errorf("kafka brokers are required")
		}
		if config.Kafka.GroupID == "" {
			return fmt.Errorf("kafka group ID is required")
		}
	}

	// Validate logger config
	validLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLevels {
		if config.Logger.Level == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("logger level must be one of: %s", strings.Join(validLevels, ", "))
	}

	validFormats := []string{"json", "text"}
	validFormat := false
	for _, format := range validFormats {
		if config.Logger.Format == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("logger format must be one of: %s", strings.Join(validFormats, ", "))
	}

	// Validate blockchain config (only when enabled)
	if config.Blockchain.Enabled {
		if config.Blockchain.NetworkURL == "" {
			return fmt.Errorf("blockchain network URL is required")
		}
		if config.Blockchain.NetworkName == "" {
			return fmt.Errorf("blockchain network name is required")
		}
	}

	return nil
}

// WatchConfig enables configuration hot reload
func WatchConfig(callback func(*Config)) error {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)

		config := &Config{}
		if err := viper.Unmarshal(config); err != nil {
			fmt.Printf("Error reloading config: %v\n", err)
			return
		}

		if err := validate(config); err != nil {
			fmt.Printf("Config validation failed after reload: %v\n", err)
			return
		}

		if callback != nil {
			callback(config)
		}
	})

	return nil
}
