package pocketbase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRecord(t *testing.T) {
	// Setup test instance
	dataDir := "../testdata/pb_data"
	inst := NewWithDataDir(dataDir).
		WithSuUserName("test@example.com").
		WithSuPassword("testpassword").
		WithListen("127.0.0.1:8090").
		WithDefaultTtl(30)

	// Start the instance in a goroutine since it blocks
	go func() {
		err := inst.Start()
		require.NoError(t, err)
	}()

	inst.WaitForReady()

	// Test cases
	tests := []struct {
		name     string
		zone     string
		record   string
		types    []string
		expected []*Record
		err      error
	}{
		{
			name:   "find single record",
			zone:   "example.com.",
			record: "cname",
			types:  []string{"CNAME"},
			expected: []*Record{
				{
					Zone:       "example.com.",
					Name:       "cname",
					RecordType: "CNAME",
					Ttl:        0,
					Content:    `{"host":"b.example.com."}`,
				},
			},
			err: nil,
		},
		{
			name:   "find multiple record types",
			zone:   "example.com.",
			record: "b",
			types:  []string{"A", "TXT"},
			expected: []*Record{
				{
					Zone:       "example.com.",
					Name:       "b",
					RecordType: "A",
					Ttl:        0,
					Content:    `{"ip":"1.1.1.1"}`,
				},
				{
					Zone:       "example.com.",
					Name:       "b",
					RecordType: "TXT",
					Ttl:        0,
					Content:    `{"text":"hello"}`,
				},
			},
			err: nil,
		},
		{
			name:     "non-existent record",
			zone:     "example.com.",
			record:   "nonexistent",
			types:    []string{"A"},
			expected: []*Record{},
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Insert test data into PocketBase before running tests
			// This would require setting up the collection and inserting records

			recs, err := inst.findRecords(tt.zone, tt.record, tt.types...)
			if tt.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(recs))

				// Compare each record
				for i, rec := range recs {
					assert.Equal(t, tt.expected[i].Zone, rec.Zone)
					assert.Equal(t, tt.expected[i].Name, rec.Name)
					assert.Equal(t, tt.expected[i].RecordType, rec.RecordType)
					assert.Equal(t, tt.expected[i].Ttl, rec.Ttl)
					assert.Equal(t, tt.expected[i].Content, rec.Content)
				}
			}
		})
	}
}

func TestFindZones(t *testing.T) {
	// Create a test instance
	inst := NewWithDataDir("../testdata/pb_data").
		WithSuUserName("test@example.com").
		WithSuPassword("testpassword").
		WithListen("127.0.0.1:8080").
		WithDefaultTtl(30)

	// Start PocketBase in a goroutine
	go func() {
		if err := inst.Start(); err != nil {
			t.Fatalf("Failed to start PocketBase: %v", err)
		}
	}()

	// Wait for PocketBase to be ready
	inst.WaitForReady()

	// Test cases
	tests := []struct {
		name          string
		expectedZones []string
		expectError   bool
	}{
		{
			name:          "No records",
			expectedZones: []string{"example.com."},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test findZones
			zones, err := inst.findZones()
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("findZones failed: %v", err)
			}

			// Check if we got the expected zones
			if len(zones) != len(tt.expectedZones) {
				t.Errorf("Expected %d zones, got %d", len(tt.expectedZones), len(zones))
			}

			// Create a map for easier comparison
			expectedMap := make(map[string]bool)
			for _, zone := range tt.expectedZones {
				expectedMap[zone] = true
			}

			for _, zone := range zones {
				if !expectedMap[zone] {
					t.Errorf("Unexpected zone found: %s", zone)
				}
			}
		})
	}
}
