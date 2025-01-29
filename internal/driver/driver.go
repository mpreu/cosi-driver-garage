package driver

import (
	cosi "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/mpreu/cosi-driver-garage/internal/client"
)

// New returns implementations for the COSI.IdentityServer and
// cosi.ProvisionerServer interfaces.
func New(driverName string, c client.ClientWithResponsesInterface) (cosi.IdentityServer, cosi.ProvisionerServer) {
	is := &identityServer{
		driverName: driverName,
	}

	ps := &provisionerServer{
		client: c,
	}

	return is, ps
}
