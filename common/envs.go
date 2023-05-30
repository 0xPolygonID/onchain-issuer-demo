package common

import (
	"errors"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	defaultPathToResolverSettings      = "./resolvers.settings.yaml"
	defaultPathToAuthV2VerificationKey = "./keys/authV2.json"
)

// ResolverSettings represent settings for resolver.
type resolverSettings map[string]struct {
	ContractAddress string `yaml:"contractAddress"`
	NetworkURL      string `yaml:"networkURL"`
	ContractOwner   string `yaml:"contractOwner"`
	ContractState   string `yaml:"contractState"`
}

func (r resolverSettings) Verify() error {
	for _, settings := range r {
		if settings.ContractAddress == "" {
			return errors.New("contract address is not set")
		}
		if settings.NetworkURL == "" {
			return errors.New("network url is not set")
		}
		if settings.ContractOwner == "" {
			return errors.New("contract owner is not set")
		}
		if settings.ContractState == "" {
			return errors.New("contract state is not set")
		}
	}
	return nil
}

var (
	ExternalServerHost        string
	InternalServerPort        string
	MongoDBHost               string
	AuthV2VerificationKeyPath string
	ResolverSettings          resolverSettings
)

func init() {
	ExternalServerHost = os.Getenv("EXTERNAL_SERVER_HOST")
	if ExternalServerHost == "" {
		panic("SERVER_HOST env variable is not set")
	}
	InternalServerPort = os.Getenv("INTERNAL_SERVER_PORT")
	if InternalServerPort == "" {
		InternalServerPort = "3333"
	}
	MongoDBHost = os.Getenv("MONGODB_HOST")
	if MongoDBHost == "" {
		MongoDBHost = "mongodb://localhost:27017/credentials"
	}
	AuthV2VerificationKeyPath = os.Getenv("AUTH_V2_VERIFICATION_KEY_PATH")
	if AuthV2VerificationKeyPath == "" {
		AuthV2VerificationKeyPath = defaultPathToAuthV2VerificationKey
	}

	resolverSettingsConfigPath := os.Getenv("RESOLVER_SETTINGS_CONFIG_PATH")
	if resolverSettingsConfigPath == "" {
		resolverSettingsConfigPath = defaultPathToResolverSettings
	}
	if err := readResolverConfig(resolverSettingsConfigPath); err != nil {
		log.Fatalf("failed read network config by path %s: %v", resolverSettingsConfigPath, err)
	}
}

func readResolverConfig(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfgs resolverSettings
	if err = yaml.Unmarshal(content, &cfgs); err != nil {
		return err
	}
	if err = cfgs.Verify(); err != nil {
		return err
	}
	ResolverSettings = cfgs
	return nil
}
