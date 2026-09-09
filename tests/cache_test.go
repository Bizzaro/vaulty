package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/declan-whiting/vaulty/internal/cache"
	"github.com/declan-whiting/vaulty/internal/models"
)

// azErrorText is the kind of payload the az cli emits when a token expires.
// It was previously written into the cache verbatim and then fatally failed
// to parse on the next startup.
const azErrorText = "ERROR: AADSTS70043: The refresh token has expired or is invalid due to sign-in frequency checks by conditional access."

type stubConfiguration struct {
	vaults []models.KeyvaultConfiguration
}

func (s stubConfiguration) GetConfiguration() models.ConfigurationList {
	return models.ConfigurationList{Keyvaults: s.vaults}
}

func newCacheServiceInTempDir(t *testing.T, vaultName string) *cache.CacheService {
	t.Helper()
	t.Chdir(t.TempDir())

	cs := cache.NewCacheService(stubConfiguration{
		vaults: []models.KeyvaultConfiguration{{Name: vaultName, SubscriptionId: "sub-id"}},
	})
	cs.EnsureCache()

	return cs
}

func TestReadKeyvaultsDiscardsUnparsableCache(t *testing.T) {
	cs := newCacheServiceInTempDir(t, "my-kv")
	path := filepath.Join("bin/cache", "my-kv-kv.json")

	if err := os.WriteFile(path, []byte(azErrorText), 0644); err != nil {
		t.Fatalf("failed to seed cache: %v", err)
	}

	if got := cs.ReadKeyvaults(); got != nil {
		t.Fatalf("expected a cache miss for unparsable data, got %#v", got)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected the unparsable keyvault cache file to be discarded")
	}
}

func TestReadSecretsDiscardsUnparsableCache(t *testing.T) {
	cs := newCacheServiceInTempDir(t, "my-kv")
	path := filepath.Join("bin/cache", "my-kv-secrets.json")

	if err := os.WriteFile(path, []byte(azErrorText), 0644); err != nil {
		t.Fatalf("failed to seed cache: %v", err)
	}

	if got := cs.ReadSecrets("my-kv"); got != nil {
		t.Fatalf("expected a cache miss for unparsable data, got %#v", got)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected the unparsable secrets cache file to be discarded")
	}
}

func TestReadKeyvaultsReturnsCachedValues(t *testing.T) {
	cs := newCacheServiceInTempDir(t, "my-kv")

	cs.WriteKeyvault("my-kv", []byte(`{"name":"my-kv","location":"canadacentral"}`))

	vaults := cs.ReadKeyvaults()
	if len(vaults) != 1 {
		t.Fatalf("expected 1 cached vault, got %d", len(vaults))
	}
	if vaults[0].Name != "my-kv" {
		t.Errorf("expected name %q, got %q", "my-kv", vaults[0].Name)
	}
	if vaults[0].SubscriptionId != "sub-id" {
		t.Errorf("expected subscription %q, got %q", "sub-id", vaults[0].SubscriptionId)
	}
}

func TestWritesIgnoreEmptyVaultName(t *testing.T) {
	cs := newCacheServiceInTempDir(t, "my-kv")

	cs.WriteKeyvault("", []byte("{}"))
	cs.WriteSecrets("", []byte("[]"))

	entries, err := os.ReadDir("bin/cache")
	if err != nil {
		t.Fatalf("failed to read cache dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no cache files for an empty vault name, got %d", len(entries))
	}
}

func TestReadLastSyncWithoutCache(t *testing.T) {
	cs := newCacheServiceInTempDir(t, "my-kv")

	if got := cs.ReadLastSync(); got == "" {
		t.Error("expected a placeholder when no sync marker exists")
	}
}

func TestClearRemovesCachedFiles(t *testing.T) {
	cs := newCacheServiceInTempDir(t, "my-kv")

	cs.WriteKeyvault("my-kv", []byte(`{"name":"my-kv"}`))
	cs.WriteSecrets("my-kv", []byte(`[{"name":"secret"}]`))
	cs.WriteLastSync([]byte("Last Sync: now"))

	if err := cs.Clear(); err != nil {
		t.Fatalf("Clear returned an error: %v", err)
	}

	entries, err := os.ReadDir("bin/cache")
	if err != nil {
		t.Fatalf("expected the cache dir to be recreated: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected an empty cache dir, got %d entries", len(entries))
	}
}
