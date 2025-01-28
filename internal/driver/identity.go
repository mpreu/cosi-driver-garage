package driver

import (
	"context"

	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

// Interface assert.
var _ cosi.IdentityServer = &identityServer{}

// identityServer implements cosi.IdentityServer.
type identityServer struct{}

// DriverGetInfo implements cosi.IdentityServer.
func (i *identityServer) DriverGetInfo(context.Context, *cosi.DriverGetInfoRequest) (*cosi.DriverGetInfoResponse, error) {
	panic("unimplemented")
}
