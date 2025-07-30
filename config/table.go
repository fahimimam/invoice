package config

import (
	"github.com/spf13/viper"
	"log"
	"sync"
)

// Table holds table configurations
type Table struct {
	Invoice string `yaml:"invoice"`
}

var tableOnce = sync.Once{}
var tableConfig *Table

// Table loads config from path
func loadTable(fileName string) error {
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.AutomaticEnv()

	tableConfig = &Table{
		Invoice: viper.GetString("table.invoice"),
	}

	log.Println("table config ", tableConfig)

	return nil
}

// GetTable returns table config
func GetTable(fileName string) *Table {
	tableOnce.Do(func() {
		err := loadTable(fileName)
		if err != nil {
			log.Fatalf("unable to read config file: %v", err)
		}
	})

	return tableConfig
}
