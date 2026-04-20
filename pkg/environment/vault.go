package environment

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// vaultConfig holds the resolved Vault connection settings.
type vaultConfig struct {
	addr       string
	token      string
	skipVerify bool
	kvPaths    []string
}

// VaultProvider fetches secrets from a HashiCorp Vault KV v2 store.
//
// Configuration is read from ~/.cagent-vault.env (preferred — no env vars needed) or
// from the standard Vault environment variables as a fallback:
//
//	VAULT_ADDR        - Vault address (e.g. https://192.168.0.251:8200)
//	VAULT_TOKEN       - Vault token for authentication
//	VAULT_SKIP_VERIFY - Set to "1" or "true" to skip TLS certificate verification
//	VAULT_KV_PATHS    - Comma-separated KV v2 mount/path pairs to read
//	                    (default: "secret/api-keys,secret/deploy-env")
//
// ~/.cagent-vault.env uses simple KEY=VALUE lines (no shell quoting or $env: prefix):
//
//	VAULT_ADDR=https://192.168.0.251:8200
//	VAULT_TOKEN=hvs.XXXXXXXXXX
//	VAULT_SKIP_VERIFY=1
//	VAULT_KV_PATHS=secret/api-keys,secret/deploy-env
//
// Values from earlier paths take precedence when the same key exists in multiple paths.
// All secrets are loaded once on first use and then cached for the lifetime of the provider.
type VaultProvider struct {
	once   sync.Once
	cache  map[string]string
	loadFn func() (map[string]string, error)
}

// vaultConfigFile is the path to the optional Vault config file.
// Populated from the user's home directory at startup.
const vaultConfigFileName = ".cagent-vault.env"

// loadVaultConfig resolves Vault connection settings by merging
// ~/.cagent-vault.env (higher priority) with OS environment variables (fallback).
// Returns nil when addr or token cannot be determined.
func loadVaultConfig() *vaultConfig {
	// Start with OS env vars as the baseline.
	cfg := map[string]string{
		"VAULT_ADDR":        os.Getenv("VAULT_ADDR"),
		"VAULT_TOKEN":       os.Getenv("VAULT_TOKEN"),
		"VAULT_SKIP_VERIFY": os.Getenv("VAULT_SKIP_VERIFY"),
		"VAULT_KV_PATHS":    os.Getenv("VAULT_KV_PATHS"),
	}

	// ~/.cagent-vault.env overrides env vars (file values take precedence so
	// the token is never in the process environment).
	if home, err := os.UserHomeDir(); err == nil {
		cfgFile := filepath.Join(home, vaultConfigFileName)
		if fileVars, err := parseSimpleEnvFile(cfgFile); err == nil {
			for k, v := range fileVars {
				if v != "" {
					cfg[k] = v
				}
			}
		}
	}

	addr := cfg["VAULT_ADDR"]
	token := cfg["VAULT_TOKEN"]
	if addr == "" || token == "" {
		return nil
	}

	rawPaths := cfg["VAULT_KV_PATHS"]
	if rawPaths == "" {
		rawPaths = "secret/api-keys,secret/deploy-env"
	}

	skipVerify := cfg["VAULT_SKIP_VERIFY"] == "1" || strings.EqualFold(cfg["VAULT_SKIP_VERIFY"], "true")

	return &vaultConfig{
		addr:       addr,
		token:      token,
		skipVerify: skipVerify,
		kvPaths:    splitTrimmed(rawPaths, ","),
	}
}

// parseSimpleEnvFile reads a KEY=VALUE file (comments and blank lines ignored).
// Values are not shell-unquoted — they are used as-is after stripping inline comments.
func parseSimpleEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		// Strip inline comments (e.g. VALUE  # comment)
		if idx := strings.Index(v, " #"); idx >= 0 {
			v = v[:idx]
		}
		result[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return result, scanner.Err()
}

// NewVaultProvider returns a VaultProvider configured from ~/.cagent-vault.env
// or Vault environment variables. Returns nil when addr or token is unavailable.
func NewVaultProvider() *VaultProvider {
	cfg := loadVaultConfig()
	if cfg == nil {
		return nil
	}

	transport := http.DefaultTransport
	if cfg.skipVerify {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // user-opted-in via VAULT_SKIP_VERIFY
		}
	}
	httpClient := &http.Client{Transport: transport}

	return &VaultProvider{
		loadFn: func() (map[string]string, error) {
			return loadVaultSecrets(httpClient, cfg.addr, cfg.token, cfg.kvPaths)
		},
	}
}

// Get implements environment.Provider. Returns the secret value by name from the
// cached Vault secrets. Vault is contacted only on the first call.
func (p *VaultProvider) Get(_ context.Context, name string) (string, bool) {
	p.once.Do(func() {
		secrets, err := p.loadFn()
		if err != nil {
			slog.Debug("VaultProvider: failed to load secrets", "error", err)
			secrets = map[string]string{}
		}
		p.cache = secrets
		slog.Debug("VaultProvider: loaded secrets from Vault", "count", len(p.cache))
	})

	v, ok := p.cache[name]
	return v, ok
}

// loadVaultSecrets reads each KV v2 path and merges the results.
// Earlier paths have higher precedence — a key from the first path will not
// be overwritten by the same key in a later path.
func loadVaultSecrets(client *http.Client, addr, token string, kvPaths []string) (map[string]string, error) {
	merged := map[string]string{}

	for _, kvPath := range kvPaths {
		secrets, err := readKVv2(client, addr, token, kvPath)
		if err != nil {
			slog.Debug("VaultProvider: could not read path, skipping", "path", kvPath, "error", err)
			continue
		}
		// Earlier paths win — only add keys not already present.
		for k, v := range secrets {
			if _, exists := merged[k]; !exists {
				merged[k] = v
			}
		}
	}

	return merged, nil
}

// kvv2Response is the JSON envelope returned by the Vault KV v2 API.
type kvv2Response struct {
	Data struct {
		Data map[string]any `json:"data"`
	} `json:"data"`
}

// readKVv2 performs a GET against the Vault KV v2 API for the given path and
// returns the key/value pairs stored there. The path must be in the form
// "<mount>/<secret-name>" (e.g. "secret/api-keys"), which gets rewritten to
// "/v1/<mount>/data/<secret-name>" internally.
func readKVv2(client *http.Client, addr, token, path string) (map[string]string, error) {
	// Rewrite "mount/name" → "/v1/mount/data/name" (KV v2 data endpoint).
	apiPath := rewriteKVv2Path(path)
	url := strings.TrimRight(addr, "/") + apiPath

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("vault: building request for %q: %w", path, err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault: GET %q: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("vault: reading body for %q: %w", path, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// Path doesn't exist — not an error, just empty.
		return map[string]string{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault: GET %q returned HTTP %d: %s", path, resp.StatusCode, truncate(string(body), 200))
	}

	var envelope kvv2Response
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("vault: decoding response for %q: %w", path, err)
	}

	result := make(map[string]string, len(envelope.Data.Data))
	for k, v := range envelope.Data.Data {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result, nil
}

// rewriteKVv2Path converts "mount/name" to "/v1/mount/data/name".
func rewriteKVv2Path(path string) string {
	// Split at the first slash to separate mount from the rest of the path.
	mount, rest, found := strings.Cut(path, "/")
	if !found {
		// No slash — treat the whole thing as the mount with empty data path.
		return "/v1/" + path + "/data/"
	}
	return "/v1/" + mount + "/data/" + rest
}

func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := parts[:0]
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
