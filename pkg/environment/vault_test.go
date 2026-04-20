package environment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func kvv2Handler(t *testing.T, secrets map[string]map[string]string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("X-Vault-Token"), "vault token header must be present")

		// Path: /v1/<mount>/data/<name>
		// Find the matching secret name from the path.
		for path, data := range secrets {
			apiPath := rewriteKVv2Path(path)
			if r.URL.Path == apiPath {
				w.Header().Set("Content-Type", "application/json")
				body := map[string]any{
					"data": map[string]any{
						"data": data,
					},
				}
				require.NoError(t, json.NewEncoder(w).Encode(body))
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}
}

func TestVaultProvider_Get(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(kvv2Handler(t, map[string]map[string]string{
		"secret/api-keys": {
			"OPENAI_API_KEY":    "sk-openai-test",
			"ANTHROPIC_API_KEY": "sk-ant-test",
		},
		"secret/deploy-env": {
			"ANTHROPIC_API_KEY": "sk-ant-SHOULD-NOT-WIN", // earlier path should win
			"MISTRAL_API_KEY":   "mistral-test",
		},
	}))
	defer server.Close()

	provider := &VaultProvider{
		loadFn: func() (map[string]string, error) {
			return loadVaultSecrets(server.Client(), server.URL, "test-token", []string{
				"secret/api-keys",
				"secret/deploy-env",
			})
		},
	}

	val, ok := provider.Get(t.Context(), "OPENAI_API_KEY")
	require.True(t, ok)
	assert.Equal(t, "sk-openai-test", val)

	val, ok = provider.Get(t.Context(), "ANTHROPIC_API_KEY")
	require.True(t, ok)
	assert.Equal(t, "sk-ant-test", val, "earlier path takes precedence")

	val, ok = provider.Get(t.Context(), "MISTRAL_API_KEY")
	require.True(t, ok)
	assert.Equal(t, "mistral-test", val, "key from second path should be present")

	_, ok = provider.Get(t.Context(), "MISSING_KEY")
	assert.False(t, ok)
}

func TestVaultProvider_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	provider := &VaultProvider{
		loadFn: func() (map[string]string, error) {
			return loadVaultSecrets(server.Client(), server.URL, "test-token", []string{"secret/missing"})
		},
	}

	_, ok := provider.Get(t.Context(), "ANYTHING")
	assert.False(t, ok)
}

func TestVaultProvider_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := &VaultProvider{
		loadFn: func() (map[string]string, error) {
			return loadVaultSecrets(server.Client(), server.URL, "test-token", []string{"secret/api-keys"})
		},
	}

	// Should gracefully return false, not panic
	_, ok := provider.Get(t.Context(), "OPENAI_API_KEY")
	assert.False(t, ok)
}

func TestRewriteKVv2Path(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"secret/api-keys", "/v1/secret/data/api-keys"},
		{"secret/deploy-env", "/v1/secret/data/deploy-env"},
		{"kv/nested/path", "/v1/kv/data/nested/path"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, rewriteKVv2Path(tt.input))
		})
	}
}

func TestNewVaultProvider_ReturnsNilWhenUnconfigured(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv
	// Clear env vars AND point home dir somewhere with no config file so
	// NewVaultProvider returns nil.
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")
	t.Setenv("HOME", t.TempDir())        // Unix
	t.Setenv("USERPROFILE", t.TempDir()) // Windows
	assert.Nil(t, NewVaultProvider())
}

func TestParseSimpleEnvFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".cagent-vault.env")
	content := `# Vault config
VAULT_ADDR=https://192.168.0.251:8200
VAULT_TOKEN=hvs.TESTTOKEN
VAULT_SKIP_VERIFY=1
VAULT_KV_PATHS=secret/api-keys,secret/deploy-env
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o600))

	got, err := parseSimpleEnvFile(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "https://192.168.0.251:8200", got["VAULT_ADDR"])
	assert.Equal(t, "hvs.TESTTOKEN", got["VAULT_TOKEN"])
	assert.Equal(t, "1", got["VAULT_SKIP_VERIFY"])
	assert.Equal(t, "secret/api-keys,secret/deploy-env", got["VAULT_KV_PATHS"])
}
