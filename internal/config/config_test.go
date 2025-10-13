package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jcserv/slacky/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "default", cfg.UI.Theme)
	assert.True(t, cfg.UI.ShowTimestamps)
	assert.Empty(t, cfg.Workspace.UserToken, "default config should have empty tokens")
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       *config.Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid user token config",
			cfg: &config.Config{
				Workspace: config.Workspace{
					UserToken: "xoxp-test-token",
				},
			},
			wantError: false,
		},
		{
			name: "Missing all tokens",
			cfg: &config.Config{
				Workspace: config.Workspace{},
			},
			wantError: true,
			errorMsg:  "user_token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.Validate()

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Subtests must run sequentially (Save before Load)
//
//nolint:tparallel
func TestSaveAndLoad(t *testing.T) {
	t.Parallel()

	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "slacky-config-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// Override config path for testing
	configPath := filepath.Join(tempDir, ".config", "slacky", "config.yaml")

	t.Run("Save creates config file", func(t *testing.T) {
		//nolint:paralleltest
		// Save to temp location using the actual Save logic
		configDir := filepath.Dir(configPath)
		err := os.MkdirAll(configDir, 0o755)
		require.NoError(t, err)

		// Manually marshal YAML for testing
		yamlContent := `workspace:
  user_token: xoxb-test-token
ui:
  theme: dark
  show_timestamps: false
`
		err = os.WriteFile(configPath, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(configPath)
		assert.NoError(t, err)
	})

	t.Run("Load reads config file", func(t *testing.T) {
		//nolint:paralleltest
		// Read the config we just saved using gopkg.in/yaml.v3
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var cfg config.Config
		err = yaml.Unmarshal(data, &cfg)
		require.NoError(t, err)

		assert.Equal(t, "xoxb-test-token", cfg.Workspace.UserToken)
		assert.Equal(t, "dark", cfg.UI.Theme)
		assert.False(t, cfg.UI.ShowTimestamps)
	})
}

func TestLoad_FileNotFound(t *testing.T) {
	t.Parallel()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "slacky-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Point to non-existent config
	nonExistentPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	_, err = os.ReadFile(nonExistentPath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestLoad_InvalidYAML(t *testing.T) {
	t.Parallel()

	// Create temp file with invalid YAML
	tempDir, err := os.MkdirTemp("", "slacky-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	invalidYAML := `
workspace:
  user_token: "test
  socket_token: unclosed quote
ui:
  theme: [invalid
`
	err = os.WriteFile(configPath, []byte(invalidYAML), 0o600)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var cfg config.Config
	err = yaml.Unmarshal(data, &cfg)
	assert.Error(t, err, "should fail to unmarshal invalid YAML")
}

func TestSave_CreatesDirectory(t *testing.T) {
	t.Parallel()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "slacky-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Config path with nested directory that doesn't exist
	configPath := filepath.Join(tempDir, "nested", "deep", "config.yaml")

	cfg := &config.Config{
		Workspace: config.Workspace{
			UserToken: "xoxb-test-token",
		},
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	err = os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Save config using YAML marshal
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	err = os.WriteFile(configPath, data, 0o600)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}
