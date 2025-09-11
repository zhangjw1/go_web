package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		envVars    map[string]string
		wantErr    bool
	}{
		{
			name:       "load default config",
			configPath: "../../configs/config.yaml",
			wantErr:    false,
		},
		{
			name:       "load dev config",
			configPath: "../../configs/config.dev.yaml",
			wantErr:    false,
		},
		{
			name:       "config file not found - use defaults",
			configPath: "nonexistent.yaml",
			wantErr:    false,
		},
		{
			name:       "override with environment variables",
			configPath: "../../configs/config.yaml",
			envVars: map[string]string{
				"SERVER_PORT": "9090",
				"SERVER_MODE": "release",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			config, err := Load(tt.configPath)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, config)

			// Check if environment variables override config
			if tt.envVars != nil {
				if port, exists := tt.envVars["SERVER_PORT"]; exists {
					assert.Equal(t, port, config.Server.Port)
				}
				if mode, exists := tt.envVars["SERVER_MODE"]; exists {
					assert.Equal(t, mode, config.Server.Mode)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Port: "8080",
					Mode: "debug",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Username: "root",
					Database: "test",
				},
				Redis: RedisConfig{
					Host: "localhost",
				},
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					GroupID: "test",
				},
				Logger: LoggerConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr: false,
		},
		{
			name: "missing server port",
			config: &Config{
				Server: ServerConfig{
					Mode: "debug",
				},
			},
			wantErr: true,
			errMsg:  "server port is required",
		},
		{
			name: "invalid server mode",
			config: &Config{
				Server: ServerConfig{
					Port: "8080",
					Mode: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "server mode must be one of",
		},
		{
			name: "missing database host",
			config: &Config{
				Server: ServerConfig{
					Port: "8080",
					Mode: "debug",
				},
				Database: DatabaseConfig{
					Username: "root",
					Database: "test",
				},
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "invalid logger level",
			config: &Config{
				Server: ServerConfig{
					Port: "8080",
					Mode: "debug",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Username: "root",
					Database: "test",
				},
				Redis: RedisConfig{
					Host: "localhost",
				},
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					GroupID: "test",
				},
				Logger: LoggerConfig{
					Level:  "invalid",
					Format: "json",
				},
			},
			wantErr: true,
			errMsg:  "logger level must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	// Clear any existing config
	os.Clearenv()

	setDefaults()

	// Test that defaults are set correctly
	config, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "debug", config.Server.Mode)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 3306, config.Database.Port)
	assert.Equal(t, "localhost", config.Redis.Host)
	assert.Equal(t, 6379, config.Redis.Port)
	assert.Equal(t, []string{"localhost:9092"}, config.Kafka.Brokers)
	assert.Equal(t, "go-web-starter", config.Kafka.GroupID)
	assert.Equal(t, "info", config.Logger.Level)
	assert.Equal(t, "json", config.Logger.Format)
}