package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/declan-whiting/vaulty/internal/models"
)

const cacheDir = "bin/cache"

type ConfigurationService interface {
	GetConfiguration() models.ConfigurationList
}

type CacheService struct {
	config ConfigurationService
}

func NewCacheService(config ConfigurationService) *CacheService {
	cs := new(CacheService)
	cs.config = config
	return cs
}

func (cs *CacheService) getKeyvaultFilePath(name string) string {
	path := filepath.Join(cacheDir, name+"-kv.json")
	return path
}

func (cs *CacheService) getSecretsFilePath(name string) string {
	path := filepath.Join(cacheDir, name+"-secrets.json")
	return path
}

func (cs *CacheService) getLastSyncPath() string {
	path := filepath.Join(cacheDir, "lastsync.txt")
	return path
}

// write persists cache contents, treating any failure as a warning.
// A cache that cannot be written is a degraded experience, not a fatal one.
func (cs *CacheService) write(path string, contents []byte) {
	cs.EnsureCache()
	if err := os.WriteFile(path, contents, 0644); err != nil {
		fmt.Printf("Failed to write cache file %s: %v\n", path, err)
	}
}

// discard removes a cache file that could not be parsed, so the next run
// refetches it from Azure instead of failing again on the same bad data.
func (cs *CacheService) discard(path string) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Failed to remove unreadable cache file %s: %v\n", path, err)
	}
}

// Clear removes every cached keyvault, secret list and sync marker.
func (cs *CacheService) Clear() error {
	if err := os.RemoveAll(filepath.Join(".", cacheDir)); err != nil {
		return err
	}
	cs.EnsureCache()
	return nil
}

func (cs *CacheService) WriteLastSync(contents []byte) {
	cs.write(cs.getLastSyncPath(), contents)
}

func (cs *CacheService) WriteKeyvault(name string, contents []byte) {
	if name == "" {
		return
	}
	cs.write(cs.getKeyvaultFilePath(name), contents)
}

func (cs *CacheService) WriteSecrets(name string, contents []byte) {
	if name == "" {
		return
	}
	cs.write(cs.getSecretsFilePath(name), contents)
}

func (cs *CacheService) EnsureCache() {
	path := filepath.Join(".", cacheDir)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create cache directory %s: %v\n", path, err)
	}
}

func (cs *CacheService) ReadKeyvaults() []models.KeyvaultModel {
	cs.EnsureCache()
	config := cs.config.GetConfiguration()
	var cachedVaults []models.KeyvaultModel

	for i, v := range config.Keyvaults {
		path := cs.getKeyvaultFilePath(v.Name)

		out, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var vault models.KeyvaultModel
		if err := json.Unmarshal(out, &vault); err != nil {
			fmt.Printf("Cached keyvault %s is unreadable (%v)\nDiscarding it and refreshing from Azure.\n", v.Name, err)
			cs.discard(path)
			return nil
		}

		vault.SubscriptionId = config.Keyvaults[i].SubscriptionId
		cachedVaults = append(cachedVaults, vault)
	}

	return cachedVaults
}

func (cs *CacheService) ReadSecrets(keyvaultName string) []models.SecretModel {
	path := cs.getSecretsFilePath(keyvaultName)

	out, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var secrets []models.SecretModel
	if err := json.Unmarshal(out, &secrets); err != nil {
		fmt.Printf("Cached secrets for %s are unreadable (%v)\nDiscarding them and refreshing from Azure.\n", keyvaultName, err)
		cs.discard(path)
		return nil
	}

	return secrets
}

func (cs *CacheService) ReadLastSync() string {
	out, err := os.ReadFile(cs.getLastSyncPath())
	if err != nil {
		return "Last Sync: never"
	}

	return string(out)
}
