package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		envVars    map[string]string
		wantErr    bool
	}{
		{
			name:       "load with defaults",
			configPath: "",
			wantErr:    false,
		},
		{
			name:       "load with environment variables",
			configPath: "",
			envVars: map[string]string{
				"PYAIRTABLE_SERVER_PORT":         "9000",
				"PYAIRTABLE_DATABASE_HOST":       "localhost",
				"PYAIRTABLE_DATABASE_PASSWORD":   "testpass",
				"PYAIRTABLE_AUTH_JWT_SECRET":     "testsecret",
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if config == nil {
					t.Error("Load() returned nil config")
					return
				}

				// Test default values
				if config.Server.Host != "localhost" {
					t.Errorf("Expected server host to be 'localhost', got %s", config.Server.Host)
				}

				// Test environment variable override
				if tt.envVars["PYAIRTABLE_SERVER_PORT"] != "" {
					expectedPort := 9000
					if config.Server.Port != expectedPort {
						t.Errorf("Expected server port to be %d, got %d", expectedPort, config.Server.Port)
					}
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Database: DatabaseConfig{
					Password: "testpass",
				},
				Auth: AuthConfig{
					JWTSecret: "testsecret",
				},
				Server: ServerConfig{
					Port: 8080,
				},
			},
			wantErr: false,
		},
		{
			name: "missing database password",
			config: &Config{
				Auth: AuthConfig{
					JWTSecret: "testsecret",
				},
				Server: ServerConfig{
					Port: 8080,
				},
			},
			wantErr: true,
		},
		{
			name: "missing JWT secret",
			config: &Config{
				Database: DatabaseConfig{
					Password: "testpass",
				},
				Server: ServerConfig{
					Port: 8080,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Database: DatabaseConfig{
					Password: "testpass",
				},
				Auth: AuthConfig{
					JWTSecret: "testsecret",
				},
				Server: ServerConfig{
					Port: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{
			name:        "development environment",
			environment: "development",
			want:        true,
		},
		{
			name:        "Development environment (uppercase)",
			environment: "Development",
			want:        true,
		},
		{
			name:        "production environment",
			environment: "production",
			want:        false,
		},
		{
			name:        "testing environment",
			environment: "testing",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Server: ServerConfig{
					Environment: tt.environment,
				},
			}
			if got := config.IsDevelopment(); got != tt.want {
				t.Errorf("Config.IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{
			name:        "production environment",
			environment: "production",
			want:        true,
		},
		{
			name:        "Production environment (uppercase)",
			environment: "Production",
			want:        true,
		},
		{
			name:        "development environment",
			environment: "development",
			want:        false,
		},
		{
			name:        "testing environment",
			environment: "testing",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Server: ServerConfig{
					Environment: tt.environment,
				},
			}
			if got := config.IsProduction(); got != tt.want {
				t.Errorf("Config.IsProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Load("")
		if err != nil {
			b.Fatal(err)
		}
	}
}