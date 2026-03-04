package azure

import (
    "encoding/json"
    "fmt"
    "os/exec"

    "github.com/declan-whiting/vaulty/internal/models"
)

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
// Writes the response to the cache.
// Requires a keyvault name and subscription id.
// Returns a KeyvaultModel.
func (az *AzureService) AzShowKeyvault(name, subscriptionId string) models.KeyvaultModel {
    out, _ := exec.Command("az", "keyvault", "show", "--name", name, "--subscription", subscriptionId, "--output", "json").CombinedOutput()
    az.CacheService.WriteKeyvault(name, out)
    var kv models.KeyvaultModel
    kv.SubscriptionId = subscriptionId
    err := json.Unmarshal(out, &kv)
    if err != nil {
        fmt.Println("Failed to parse JSON for keyvaults")
        fmt.Println(err)
    }

    return kv
}

// Equivlant to an `az keyvault secret list` azure cli command.
// Writes the response to the cache.
// Requires a keyvault name and subscription id.
// Returns a list of SecretModels.
func (az *AzureService) AzGetSecrets(name, subscriptionId string) []models.SecretModel {
    out, _ := exec.Command("az", "keyvault", "secret", "list", "--vault-name", name, "--subscription", subscriptionId, "--output", "json").CombinedOutput()
    az.CacheService.WriteSecrets(name, out)
    var response []models.SecretModel
    err := json.Unmarshal(out, &response)
    if err != nil {
        fmt.Println("Failed to parse JSON for secrets")
        fmt.Println(err)
    }

    return response
}

// Equivlant to an `az keyvault secret list` azure cli command.
// Secrets are not cached.
// Requires a secret name, a keyvault name and subscription id.
// Returns a secret in json format as a string.
func (az *AzureService) AzShowSecret(secretName, vaultName, subscriptionId string) string {
    secret, ok := az.SecretsStow[subscriptionId+vaultName+secretName]
    if !ok {
        out, _ := exec.Command("az", "keyvault", "secret", "show", "--vault-name", vaultName, "--name", secretName, "--subscription", subscriptionId, "--output", "json").CombinedOutput()
        az.SecretsStow[subscriptionId+vaultName+secretName] = string(out)
        return string(out)
    } else {
        return secret
    }

}

// ClearSecret removes the in-memory stow entry for a secret so the next
// call to AzShowSecret fetches a fresh value from Azure.
func (az *AzureService) ClearSecret(secretName, vaultName, subscriptionId string) {
    delete(az.SecretsStow, subscriptionId+vaultName+secretName)
}
