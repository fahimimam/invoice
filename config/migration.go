package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

type InvoiceDetails struct {
	FirstName string   `mapstructure:"first_name"`
	LastName  string   `mapstructure:"last_name"`
	Email     string   `mapstructure:"email"`
	Phone     string   `mapstructure:"phone"`
	Password  string   `mapstructure:"password"`
	Roles     []string `mapstructure:"roles"`
}

// MigrationDetails holds table configurations
type MigrationDetails struct {
	Invoices []InvoiceDetails `mapstructure:"org"`
}

var migrationOnce = sync.Once{}
var migrationConfig *MigrationDetails

// loadToken loads config from path
func loadMigration(fileName string) error {
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.AutomaticEnv()

	if err := viper.UnmarshalKey("migration", &migrationConfig); err != nil {
		log.Fatalf("Failed to load migration config: %v", err)
	}

	//migrationConfig = &MigrationDetails{
	//	FirstName: viper.GetString("migration.first_name"),
	//	LastName:  viper.GetString("migration.last_name"),
	//	Email:     viper.GetString("migration.email"),
	//	Phone:     viper.GetString("migration.phone"),
	//	Password:  viper.GetString("migration.password"),
	//	Type:      viper.GetString("migration.type"),
	//	Roles:     pq.StringArray{"master"},
	//	OrgName:   viper.GetString("migration.org_name"),
	//}

	return nil
}

// GetMigration returns token config
func GetMigration(fileName string) *MigrationDetails {
	migrationOnce.Do(func() {
		err := loadMigration(fileName)
		if err != nil {
			log.Fatalf("unable to read config file: %v", err)
		}
	})

	return migrationConfig
}
