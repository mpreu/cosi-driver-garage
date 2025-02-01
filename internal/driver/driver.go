package driver

import (
	"log/slog"

	cosi "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/mpreu/cosi-driver-garage/internal/client"
	"github.com/mpreu/cosi-driver-garage/internal/config"
)

// New returns implementations for the COSI.IdentityServer and
// cosi.ProvisionerServer interfaces.
func New(driverName string, config *config.Garage, c client.ClientWithResponsesInterface, logger *slog.Logger) (cosi.IdentityServer, cosi.ProvisionerServer) {
	is := &identityServer{
		driverName: driverName,
	}

	ps := &provisionerServer{
		client: c,
		config: config,
		logger: logger,
	}

	return is, ps
}
