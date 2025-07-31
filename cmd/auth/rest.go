package cmd

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"github.com/triapex/auth/api"
	"github.com/triapex/auth/config"
	"github.com/triapex/auth/internal/infra"
	"github.com/triapex/auth/internal/infra/postgres"
	"github.com/triapex/auth/internal/infra/redis"
	"github.com/triapex/auth/internal/infra/sentry"
	"github.com/triapex/auth/internal/repo"
	"github.com/triapex/auth/internal/service"
	"github.com/triapex/auth/logger"
)

const DefaultAccessTokenDuration = 10
const DefaultRefreshTokenDuration = 30

// srvCmd is the serve sub command to start the api server
var srvCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve serves the auth server",
	RunE:  serve,
}

func init() {
	srvCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "app.config.yaml", "config file path")
}

func serve(cmd *cobra.Command, args []string) error {
	cfgApp := config.GetApp(cfgPath)
	cfgPostgres := config.GetPostgres(cfgPath)
	cfgDBTable := config.GetTable(cfgPath)
	cfgRedis := config.GetRedis(cfgPath)
	cfgSentry := config.GetSentry(cfgPath)
	cfgInvoice := config.GetInvoice(cfgPath)

	ctx := context.Background()
	lgr := logger.DefaultOutStructLogger

	// connect postgres db
	db, err := postgres.NewConnection(ctx, cfgPostgres)
	if err != nil {
		return err
	}
	defer db.Close(ctx)

	// connect redis db
	kv, err := redis.New(cfgRedis.URL, cfgRedis.RedisTimeOut, "auth")
	if err != nil {
		//return err
	}
	defer kv.Close()

	// initialize sentry
	err = sentry.NewInit(cfgSentry.URL)
	if err != nil {
		//return err
	}

	// Initialize Repositories.
	grpcConn, err := newGrpcConnection()
	if err != nil {
		log.Fatalf("Failed to initialize gRPC connection: %v", err)
	}
	defer func() {
		err = grpcConn.Close()
		log.Fatalf("Failed to close GRPC connection: %v", err)
	}()
	invoiceRepo := repo.NewInvoice(cfgDBTable, db)

	invoiceSvc := service.NewInvoice(invoiceRepo, kv, lgr)
	api.SetLogger(logger.DefaultOutLogger)

	errChan := make(chan error)
	go func() {
		if err := startHealthServer(cfgApp, db, kv); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := startApiServer(cfgApp, invoiceSvc, grpcConn, lgr, cfgInvoice); err != nil {
			errChan <- err
		}
	}()
	return <-errChan

}

func startHealthServer(cfg *config.Application, db infra.DB, kv infra.KV) error {
	log.Println("startHealthServer")
	sc := api.NewSystemController(db, kv)
	api.NewSystemRouter(sc)
	r := chi.NewMux()
	r.Mount("/system/v1", api.NewSystemRouter(sc))

	srvr := http.Server{
		Addr:    getAddressFromHostAndPort(cfg.Host, 3550),
		Handler: r,
		//ErrorLog: logger.DefaultErrLogger,
		//WriteTimeout: cfg.WriteTimeout,
		//ReadTimeout:  cfg.ReadTimeout,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	if err := srvr.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	graceful := func() error {
		log.Println("To shutdown immediately press again")

		return nil
	}

	errCh := make(chan error)
	forced := func() error {
		log.Println("Shutting down server forcefully")
		return nil
	}
	sigs := []os.Signal{syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM}

	go func() {
		errCh <- HandleSignals(sigs, graceful, forced)
	}()

	return <-errCh
}

func startApiServer(cfg *config.Application, invoiceSvc service.InvoiceService, grpc *grpc.ClientConn, lgr logger.StructLogger, invoiceConfig *config.Invoice) error {

	invoiceController := api.NewInvoiceController(invoiceSvc, grpc, lgr)
	invoiceController.SetLogger(lgr)

	r := chi.NewMux()
	r.Route("/auth/api/v1", func(rt chi.Router) {

		rt.Mount("/invoice", api.NewInvoiceRouter(invoiceController))
	})

	srvr := http.Server{
		Addr:              getAddressFromHostAndPort(cfg.Host, cfg.Port),
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	return ManageServer(&srvr, cfg.GracefulTimeout*time.Second)
}

func ManageServer(srvr *http.Server, gracePeriod time.Duration) error {
	errCh := make(chan error)

	sigs := []os.Signal{syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM, os.Interrupt}

	graceful := func() error {
		log.Println("Shutting down server gracefully...")
		log.Println("To shutdown immediately, press again.")

		ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()

		go func() {
			// Countdown in the same goroutine
			countdownDuration := int(gracePeriod.Seconds()) // Convert to seconds
			for i := countdownDuration; i > 0; i-- {
				log.Printf("Graceful shutdown in %d seconds...\n", i)
				time.Sleep(1 * time.Second)
			}
		}()

		return srvr.Shutdown(ctx)
	}

	forced := func() error {
		log.Println("Shutting down server forcefully")
		return srvr.Close()
	}

	go func() {
		log.Println("Starting server on", srvr.Addr)
		if err := srvr.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		errCh <- HandleSignals(sigs, graceful, forced)
	}()

	return <-errCh
}

// HandleSignals listen on the registered signals and fires the gracefulHandler for the
// first signal and the forceHandler (if any) for the next this function blocks and
// return any error that returned by any of the handlers first
func HandleSignals(sigs []os.Signal, gracefulHandler, forceHandler func() error) error {
	sigCh := make(chan os.Signal)
	errCh := make(chan error, 1)

	signal.Notify(sigCh, sigs...)
	defer signal.Stop(sigCh)

	grace := true
	for {
		select {
		case err := <-errCh:
			return err
		case <-sigCh:
			if grace {
				grace = false
				go func() {
					errCh <- gracefulHandler()
				}()
			} else if forceHandler != nil {
				err := forceHandler()
				errCh <- err
			}
		}
	}
}

func getAddressFromHostAndPort(host string, port int) string {
	addr := host
	if port != 0 {
		addr = addr + ":" + strconv.Itoa(port)
	}
	return addr
}
