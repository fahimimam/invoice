package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

// Sentry holds sentry config
type Sentry struct {
	URL string `yaml:"sentry.dsn"`
}

var sentryOnce = sync.Once{}
var sentryConfig *Sentry

// loadSentry loads config from path
func loadSentry(fileName string) error {
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.AutomaticEnv()

	sentryConfig = &Sentry{
		URL: viper.GetString("sentry.dsn"),
	}

	return nil
}

// GetSentry returns redis config
func GetSentry(fileName string) *Sentry {
	sentryOnce.Do(func() {
		err := loadSentry(fileName)
		if err != nil {
			log.Fatalf("unable to read config file: %v", err)
		}
	})

	return sentryConfig
}
