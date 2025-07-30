package config

import (
	"github.com/spf13/viper"
	"log"
	"sync"
)

// Identity holds table configurations
type Identity struct {
	Channel   string `yaml:"channel"`
	ChainCode string `yaml:"chain_code"`
}

var identityOnce = sync.Once{}
var identityConfig *Identity

// loadInvoice loads config from path
func loadIdentity(fileName string) error {
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.AutomaticEnv()

	identityConfig = &Identity{
		Channel: viper.GetString("smart_contract.identity.channel"),
	}

	log.Println("table config ", identityConfig)

	return nil
}

// GetIdentity returns table config
func GetIdentity(fileName string) *Identity {
	identityOnce.Do(func() {
		err := loadIdentity(fileName)
		if err != nil {
			log.Fatalf("unable to read config file: %v", err)
		}
	})

	return identityConfig
}
