package cmd

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/provisioner"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"

	"github.com/mpreu/cosi-driver-garage/internal/client"
	"github.com/mpreu/cosi-driver-garage/internal/config"
	"github.com/mpreu/cosi-driver-garage/internal/driver"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Config{
		DriverName:          getEnv("COSI_DRIVER_NAME", "garage.objectstorage.k8s.io"),
		COSIEndpoint:        getEnv("COSI_ENDPOINT", "unix:///var/lib/cosi/cosi.sock"),
		GarageAdminEndpoint: getEnv("GARAGE_ADMIN_ENDPOINT", ""),
		GarageAdminToken:    getEnv("GARAGE_ADMIN_TOKEN", ""),
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("Error validating config", "error", err)
		os.Exit(1)
	}

	if err := run(context.Background(), &cfg); err != nil {
		logger.Error("Error running the driver", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(ctx,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Setup Garage HTTP client.
	tokenProvider, err := securityprovider.NewSecurityProviderBearerToken(cfg.GarageAdminToken)
	if err != nil {
		return err
	}

	c, err := client.NewClientWithResponses(cfg.GarageAdminEndpoint,
		client.WithHTTPClient(&http.Client{}),
		client.WithRequestEditorFn(tokenProvider.Intercept),
	)
	if err != nil {
		return err
	}

	// Run COSI server.
	is, ps := driver.New(cfg.DriverName, c)

	server, err := provisioner.NewDefaultCOSIProvisionerServer(
		cfg.COSIEndpoint,
		is,
		ps,
	)
	if err != nil {
		return err
	}

	return server.Run(ctx)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
