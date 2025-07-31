package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/triapex/auth/config"
	"github.com/triapex/auth/internal/infra/postgres"
	"github.com/triapex/auth/internal/infra/redis"
	"github.com/triapex/auth/internal/repo"
	"github.com/triapex/auth/internal/service"
	"github.com/triapex/auth/logger"
	"github.com/triapex/auth/model"
	"log"
)

var invoiceSVC service.InvoiceService
var db *postgres.Postgres
var migrationConfig *config.MigrationDetails

var migrationRoot = &cobra.Command{
	Use:   "migration",
	Short: "Run database migrations",
	Long:  `Migration is a tool to generate and modify databse tables`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfgPostgres := config.GetPostgres(cfgPath)
		cfgDBTable := config.GetTable(cfgPath)
		cfgRedis := config.GetRedis(cfgPath)
		migrationConfig = config.GetMigration(cfgPath)
		ctx := context.Background()
		lgr := logger.DefaultOutStructLogger
		var err error
		// connect postgres db
		db, err = postgres.NewConnection(ctx, cfgPostgres)
		if err != nil {
			return err
		}
		//defer db.Close(ctx)
		// connect redis db

		kv, err := redis.New(cfgRedis.URL, cfgRedis.RedisTimeOut, "auth")
		if err != nil {
			return err
		}
		defer kv.Close()

		invoiceRepo := repo.NewInvoice(cfgDBTable, db)
		invoiceSVC = service.NewInvoice(invoiceRepo, kv, lgr)
		return nil
	},
}

func init() {
	migrationRoot.PersistentFlags().StringVarP(&cfgPath, "config", "c", "config.yaml", "config file path")
}

var migrationUp = &cobra.Command{
	Use:   "up",
	Short: "Populate tables in database",
	Long:  `Populate tables in database`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Populating database...")
		ctx := context.Background()
		defer db.Close(context.Background())
		if err := db.DB.AutoMigrate(model.Models...); err != nil {
			log.Println("Failed to migrate database. Error: ", err.Error())
			return err
		}

		for _, invoiceConfig := range migrationConfig.Invoices {

			log.Printf("Created organization '%s'", invoiceConfig.FirstName)
			invoice := &model.InvoiceInfo{
				//FirstName: invoiceConfig.FirstName,
				//LastName:  invoiceConfig.LastName,
				//Phone:     invoiceConfig.Phone,
				//Email:     invoiceConfig.Email,
				//Password:  invoiceConfig.Password,
				//Roles:     pq.StringArray(invoiceConfig.Roles), // Optional but dynamic if present
				//Verified:  true,
			}

			_, err := invoiceSVC.CreateInvoice(ctx, &model.CreateInvoicePayload{
				InvoiceInfo: invoice,
			})
			if err != nil {
				log.Printf("Failed to create user '%s %s' for organization '%s': %s",
					invoiceConfig.FirstName, invoiceConfig.LastName, invoiceConfig.FirstName, err)
				return err
			}

			log.Printf("Created user '%s %s' for organization '%s'",
				invoiceConfig.FirstName, invoiceConfig.LastName, invoiceConfig.FirstName)

		}

		log.Println("Database populated successfully!")
		return nil
	},
}

var migrationDown = &cobra.Command{
	Use:   "down",
	Short: "Drop tables from database",
	Long:  `Drop tables from database`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		log.Println("Dropping database table...")
		i := 1
		for _, table := range model.Models {
			fmt.Printf("Iteration: %v\n", i)
			i++
			err = db.Migrator().DropTable(table)
		}
		log.Println("Database dopped successfully!")
		return err
	},
}

func init() {
	migrationRoot.AddCommand(
		migrationUp,
		migrationDown,
	)
}
