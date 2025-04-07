package pocketbase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithDataDir(t *testing.T) {
	tempDir := t.TempDir()
	inst := NewWithDataDir(tempDir)
	assert.NotNil(t, inst)
	assert.NotNil(t, inst.pb)
	assert.Equal(t, tempDir, inst.pb.DataDir())
}

func TestWithSuUserName(t *testing.T) {
	inst := NewWithDataDir(t.TempDir())
	email := "test@example.com"
	inst = inst.WithSuUserName(email)
	assert.Equal(t, email, inst.suEmail)
}

func TestWithSuPassword(t *testing.T) {
	inst := NewWithDataDir(t.TempDir())
	password := "testpassword"
	inst = inst.WithSuPassword(password)
	assert.Equal(t, password, inst.suPassword)
}

func TestWithListen(t *testing.T) {
	inst := NewWithDataDir(t.TempDir())
	listen := "127.0.0.1:8080"
	inst = inst.WithListen(listen)
	assert.Equal(t, listen, inst.listen)
}

func TestStart(t *testing.T) {
	tempDir := t.TempDir()
	inst := NewWithDataDir(tempDir).
		WithSuUserName("test@example.com").
		WithSuPassword("testpassword").
		WithListen("127.0.0.1:8090")

	// Start the instance in a goroutine since it blocks
	go func() {
		err := inst.Start()
		assert.NoError(t, err)
	}()

	inst.WaitForReady()

	// Clean up
	defer os.RemoveAll(tempDir)
}

func TestToAbsPath(t *testing.T) {
	// Test absolute path
	absPath := "/absolute/path"
	result := toAbsPath(absPath)
	assert.Equal(t, absPath, result)

	// Test relative path
	relPath := "relative/path"
	execPath, err := os.Executable()
	require.NoError(t, err)
	execDir := filepath.Dir(execPath)
	expectedPath := filepath.Join(execDir, relPath)
	result = toAbsPath(relPath)
	assert.Equal(t, expectedPath, result)
}
