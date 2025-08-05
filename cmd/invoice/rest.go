package cmd

import (
	"context"
	"github.com/fahimimam/invoice/api"
	"github.com/fahimimam/invoice/config"
	"github.com/fahimimam/invoice/internal/network"
	"github.com/fahimimam/invoice/internal/service"
	"github.com/fahimimam/invoice/logger"
	"github.com/unidoc/unipdf/v4/common/license"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
)

const DefaultAccessTokenDuration = 10
const DefaultRefreshTokenDuration = 30

// srvCmd is the serve sub command to start the api server
var srvCmd = &cobra.Command{
	Use:     "serve",
	Short:   "serve serves the auth server",
	RunE:    serve,
	Aliases: []string{"s"},
}

func init() {

	srvCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "app.config.yaml", "config file path")
}

func serve(cmd *cobra.Command, args []string) error {
	cfgApp := config.GetApp(cfgPath)
	//cfgInvoice := config.GetInvoice(cfgPath)

	lgr := logger.DefaultOutStructLogger

	// Make sure to load your metered License API key prior to using the library.
	// If you need a key, you can sign up and create a free one at https://cloud.unidoc.io
	err := license.SetMeteredKey(cfgApp.UnidocApiKey)
	if err != nil {
		panic(err)
	}

	// Connect NewNetworkRPCConnection...
	networkConn, err := network.NewNetworkRPCConnection()
	if err != nil {
		log.Fatalf("Failed to initialize network rpc connection: %v", err)
	}
	defer func() {
		err = networkConn.NetworkConnectionClose()
		log.Fatalf("Failed to close network rpc connection: %v", err)
	}()

	invoiceSvc := service.NewInvoiceService(networkConn, lgr)
	api.SetLogger(logger.DefaultOutLogger)

	errChan := make(chan error)
	go func() {
		if err := startApiServer(cfgApp, invoiceSvc, lgr); err != nil {
			errChan <- err
		}
	}()
	return <-errChan

}

func startApiServer(cfg *config.Application, invoiceSvc service.InvoiceServiceInterface, lgr logger.StructLogger) error {

	invoiceCtlr := api.NewInvoiceController(invoiceSvc, lgr)
	invoiceCtlr.SetLogger(lgr)

	r := chi.NewMux()
	r.Mount("/invoice/api/v1", api.NewInvoiceRouter(invoiceCtlr))

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
