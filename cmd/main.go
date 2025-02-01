package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
		COSIEndpoint: getEnv("COSI_ENDPOINT", "unix:///var/lib/cosi/cosi.sock"),
		DriverName:   getEnv("X_COSI_DRIVER_NAME", "garage.objectstorage.k8s.io"),
		Garage: &config.Garage{
			Endpoint:           getEnv("GARAGE_ENDPOINT", ""),
			Region:             getEnv("GARAGE_REGION", ""),
			AdminEndpoint:      getEnv("GARAGE_ADMIN_ENDPOINT", ""),
			AdminToken:         getEnv("GARAGE_ADMIN_TOKEN", ""),
			InsecureSkipVerify: asBool(getEnv("GARAGE_INSECURE_SKIP_VERIFY", "false")),
		},
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("Error validating config", "error", err)
		os.Exit(1)
	}

	if err := run(context.Background(), &cfg, logger); err != nil {
		logger.Error("Error running the driver", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(ctx,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Setup Garage HTTP client.
	tokenProvider, err := securityprovider.NewSecurityProviderBearerToken(cfg.Garage.AdminToken)
	if err != nil {
		return err
	}

	c, err := client.NewClientWithResponses(cfg.Garage.AdminEndpoint,
		client.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.Garage.InsecureSkipVerify,
				},
			},
		}),
		client.WithRequestEditorFn(tokenProvider.Intercept),
	)
	if err != nil {
		return err
	}

	// Run COSI server.
	is, ps := driver.New(cfg.DriverName, cfg.Garage, c, logger)

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

func asBool(v string) bool {
	b, _ := strconv.ParseBool(v)
	return b
}
