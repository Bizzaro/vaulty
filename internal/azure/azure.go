package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/declan-whiting/vaulty/internal/models"
)

// runAz executes an azure cli command and returns stdout only.
// Diagnostics such as "ERROR: AADSTS70043: The refresh token has expired" are
// written to stderr, so keeping the streams apart stops that text from ever
// being mistaken for a JSON payload and persisted to the cache.
func runAz(args ...string) ([]byte, error) {
	cmd := exec.Command("az", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = err.Error()
		}
		return nil, fmt.Errorf("az %s: %s", strings.Join(args, " "), detail)
	}

	return out, nil
}

// A CacheService for the Azure package needs to be able to WriteKeyvaults and WriteSecrets to the cache.
type CacheService interface {
	WriteKeyvault(string, []byte)
	WriteSecrets(string, []byte)
}

// A AzureService writes queries to the AzureCLI using the currently logged in account.
type AzureService struct {
	CacheService CacheService
	SecretsStow  map[string]string
}

// Returns a new instance of the AzureService.
// Requires the CacheService interface.
func NewAzureService(cache CacheService) *AzureService {
	azure := new(AzureService)
	azure.CacheService = cache
	azure.SecretsStow = make(map[string]string)
	return azure
}

// Equivlant to an `az keyvault show` azure cli command.
// Writes the response to the cache only when the command succeeds and the
// payload parses, so a failed call can never poison the cache.
// Requires a keyvault name and subscription id.
// Returns a KeyvaultModel.
func (az *AzureService) AzShowKeyvault(name, subscriptionId string) models.KeyvaultModel {
	var kv models.KeyvaultModel
	kv.SubscriptionId = subscriptionId

	out, err := runAz("keyvault", "show", "--name", name, "--subscription", subscriptionId, "--output", "json")
	if err != nil {
		fmt.Printf("Failed to get keyvault %s\n%v\n", name, err)
		return kv
	}

	if err := json.Unmarshal(out, &kv); err != nil {
		fmt.Printf("Failed to parse JSON for keyvault %s\n%v\n", name, err)
		return kv
	}

	az.CacheService.WriteKeyvault(name, out)
	return kv
}

// Equivlant to an `az keyvault secret list` azure cli command.
// Writes the response to the cache only when the command succeeds and the
// payload parses, so a failed call can never poison the cache.
// Requires a keyvault name and subscription id.
// Returns a list of SecretModels.
func (az *AzureService) AzGetSecrets(name, subscriptionId string) []models.SecretModel {
	out, err := runAz("keyvault", "secret", "list", "--vault-name", name, "--subscription", subscriptionId, "--output", "json")
	if err != nil {
		fmt.Printf("Failed to get secrets for %s\n%v\n", name, err)
		return nil
	}

	var response []models.SecretModel
	if err := json.Unmarshal(out, &response); err != nil {
		fmt.Printf("Failed to parse JSON for secrets in %s\n%v\n", name, err)
		return nil
	}

	az.CacheService.WriteSecrets(name, out)
	return response
}

// Equivlant to an `az keyvault secret list` azure cli command.
// Secrets are not cached.
// Requires a secret name, a keyvault name and subscription id.
// Returns a secret in json format as a string.
func (az *AzureService) AzShowSecret(secretName, vaultName, subscriptionId string) string {
	key := subscriptionId + vaultName + secretName
	if secret, ok := az.SecretsStow[key]; ok {
		return secret
	}

	out, err := runAz("keyvault", "secret", "show", "--vault-name", vaultName, "--name", secretName, "--subscription", subscriptionId, "--output", "json")
	if err != nil {
		// Surfaced to the detail view so the user sees why it failed, but not
		// stowed: a transient failure must not stick for the rest of the session.
		return err.Error()
	}

	az.SecretsStow[key] = string(out)
	return string(out)
}

// ClearSecret removes the in-memory stow entry for a secret so the next
// call to AzShowSecret fetches a fresh value from Azure.
func (az *AzureService) ClearSecret(secretName, vaultName, subscriptionId string) {
	delete(az.SecretsStow, subscriptionId+vaultName+secretName)
}
