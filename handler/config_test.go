package handler

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config.Listen != defaultListen {
		t.Errorf("expected Listen to be %s, got %s", defaultListen, config.Listen)
	}
	if config.DataDir != defaultDataDir {
		t.Errorf("expected DataDir to be %s, got %s", defaultDataDir, config.DataDir)
	}
	if config.SuEmail != defaultSuEmail {
		t.Errorf("expected SuUserName to be %s, got %s", defaultSuEmail, config.SuEmail)
	}
	if config.SuPassword != defaultSuPassword {
		t.Errorf("expected SuPassword to be %s, got %s", defaultSuPassword, config.SuPassword)
	}
	if config.CacheCapacity != defaultCacheCapacity {
		t.Errorf("expected CacheCapacity to be %d, got %d", defaultCacheCapacity, config.CacheCapacity)
	}
	if config.DefaultTtl != defaultDefaultTtl {
		t.Errorf("expected DefaultTtl to be %d, got %d", defaultDefaultTtl, config.DefaultTtl)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  NewConfig(),
			wantErr: false,
		},
		{
			name:    "valid config2",
			config:  NewConfig().WithListen("[::]:8090"),
			wantErr: false,
		},
		{
			name:    "empty listen address",
			config:  NewConfig().WithListen(""),
			wantErr: true,
		},
		{
			name:    "invalid listen address format",
			config:  NewConfig().WithListen("invalid:address"),
			wantErr: true,
		},
		{
			name:    "empty data directory",
			config:  NewConfig().WithDataDir(""),
			wantErr: true,
		},
		{
			name:    "empty superuser email",
			config:  NewConfig().WithSuEmail(""),
			wantErr: true,
		},
		{
			name:    "empty superuser password",
			config:  NewConfig().WithSuPassword(""),
			wantErr: true,
		},
		{
			name:    "negative cache capacity",
			config:  NewConfig().WithCacheCapacity(-1),
			wantErr: true,
		},
		{
			name:    "negative default TTL",
			config:  NewConfig().WithDefaultTtl(-1),
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

func TestConfigBuilderMethods(t *testing.T) {
	config := NewConfig()

	// Test WithListen
	newListen := "127.0.0.1:8080"
	config = config.WithListen(newListen)
	if config.Listen != newListen {
		t.Errorf("WithListen() failed, expected %s, got %s", newListen, config.Listen)
	}

	// Test WithDataDir
	newDataDir := "/tmp/pb_data"
	config = config.WithDataDir(newDataDir)
	if config.DataDir != newDataDir {
		t.Errorf("WithDataDir() failed, expected %s, got %s", newDataDir, config.DataDir)
	}

	// Test WithSuEmail
	newSuUserName := "admin@example.com"
	config = config.WithSuEmail(newSuUserName)
	if config.SuEmail != newSuUserName {
		t.Errorf("WithSuEmail() failed, expected %s, got %s", newSuUserName, config.SuEmail)
	}

	// Test WithSuPassword
	newSuPassword := "newpassword"
	config = config.WithSuPassword(newSuPassword)
	if config.SuPassword != newSuPassword {
		t.Errorf("WithSuPassword() failed, expected %s, got %s", newSuPassword, config.SuPassword)
	}

	// Test WithCacheCapacity
	newCacheCapacity := 100
	config = config.WithCacheCapacity(newCacheCapacity)
	if config.CacheCapacity != newCacheCapacity {
		t.Errorf("WithCacheCapacity() failed, expected %d, got %d", newCacheCapacity, config.CacheCapacity)
	}

	// Test WithDefaultTtl
	newDefaultTtl := 60
	config = config.WithDefaultTtl(newDefaultTtl)
	if config.DefaultTtl != newDefaultTtl {
		t.Errorf("WithDefaultTtl() failed, expected %d, got %d", newDefaultTtl, config.DefaultTtl)
	}
}

func TestConfigMixWithEnv(t *testing.T) {
	config := NewConfig()

	// Test with no environment variables set
	config = config.MixWithEnv()
	if config.SuEmail != defaultSuEmail {
		t.Errorf("expected SuUserName to be %s when no env var set, got %s", defaultSuEmail, config.SuEmail)
	}
	if config.SuPassword != defaultSuPassword {
		t.Errorf("expected SuPassword to be %s when no env var set, got %s", defaultSuPassword, config.SuPassword)
	}

	// Test with environment variables set
	expectedSuUserName := "env@example.com"
	expectedSuPassword := "envpassword"
	t.Setenv("COREDNS_PB_SUPERUSER_EMAIL", expectedSuUserName)
	t.Setenv("COREDNS_PB_SUPERUSER_PWD", expectedSuPassword)

	config = NewConfig().MixWithEnv()
	if config.SuEmail != expectedSuUserName {
		t.Errorf("expected SuUserName to be %s from env var, got %s", expectedSuUserName, config.SuEmail)
	}
	if config.SuPassword != expectedSuPassword {
		t.Errorf("expected SuPassword to be %s from env var, got %s", expectedSuPassword, config.SuPassword)
	}

	// Test with partial environment variables set
	t.Setenv("COREDNS_PB_SUPERUSER_EMAIL", expectedSuUserName)
	t.Setenv("COREDNS_PB_SUPERUSER_PWD", "")

	config = NewConfig().MixWithEnv()
	if config.SuEmail != expectedSuUserName {
		t.Errorf("expected SuUserName to be %s from env var, got %s", expectedSuUserName, config.SuEmail)
	}
	if config.SuPassword != defaultSuPassword {
		t.Errorf("expected SuPassword to be %s (default) when env var empty, got %s", defaultSuPassword, config.SuPassword)
	}
}
