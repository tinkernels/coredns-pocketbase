package coredns_pocketbase

import (
	"testing"

	"github.com/coredns/caddy"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name          string
		config        string
		expectedError bool
	}{
		{
			name: "valid configuration",
			config: `pocketbase {
				listen 127.0.0.1:8090
				data_dir ./data
				su_email admin@example.com
				su_password password123
				default_ttl 3600
				cache_capacity 1000
			}`,
			expectedError: false,
		},
		{
			name: "valid configuration - invalid default_ttl but using default",
			config: `pocketbase {
				listen 127.0.0.1:8090
				data_dir ./data
				su_email admin@example.com
				su_password password123
				default_ttl invalid
				cache_capacity 1000
			}`,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := caddy.NewTestController("dns", tt.config)
			err := setup(c)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
