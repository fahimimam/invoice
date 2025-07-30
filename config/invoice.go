package config

import (
	"github.com/spf13/viper"
	"log"
	"sync"
)

// Invoice holds table configurations
type Invoice struct {
	Channel   string `yaml:"channel"`
	ChainCode string `yaml:"chain_code"`
}

var invoiceOnce = sync.Once{}
var invoiceConfig *Invoice

// loadInvoice loads config from path
func loadInvoice(fileName string) error {
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.AutomaticEnv()

	invoiceConfig = &Invoice{
		Channel: viper.GetString("smart_contract.identity.channel"),
	}

	log.Println("table config ", invoiceConfig)

	return nil
}

// GetInvoice returns table config
func GetInvoice(fileName string) *Invoice {
	invoiceOnce.Do(func() {
		err := loadInvoice(fileName)
		if err != nil {
			log.Fatalf("unable to read config file: %v", err)
		}
	})

	return invoiceConfig
}
