package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/provisioner"

	"github.com/mpreu/cosi-driver-garage/internal/config"
	"github.com/mpreu/cosi-driver-garage/internal/driver"
)

func main() {
	cfg := config.Config{
		DriverName:   getEnv("COSI_DRIVER_NAME", "garage.objectstorage.k8s.io"),
		COSIEndpoint: getEnv("COSI_ENDPOINT", "unix:///var/lib/cosi/cosi.sock"),
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

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

	is, ps := driver.New()

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
