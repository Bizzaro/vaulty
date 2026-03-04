package configuration

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/declan-whiting/vaulty/internal/models"
	"gopkg.in/yaml.v3"
)

type ConfigrationService struct{}

func NewConfigurationService() *ConfigrationService {
	return new(ConfigrationService)
}

func (cs *ConfigrationService) GetConfiguration() models.ConfigurationList {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	cacheVaultFile, err := os.Open(filepath.Join(home, ".vaulty.conf"))
	if err != nil {
		log.Fatal(err)
	}
	var vaults models.ConfigurationList
	out, _ := io.ReadAll(cacheVaultFile)

	err = yaml.Unmarshal(out, &vaults)
	if err != nil {
		log.Fatal(err)
	}

	return vaults
}
