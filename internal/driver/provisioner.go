package driver

import (
	"context"

	cosi "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/mpreu/cosi-driver-garage/internal/client"
)

// Interface assert.
var _ cosi.ProvisionerServer = &provisionerServer{}

// provisionerServer implements cosi.ProvisionerServer.
type provisionerServer struct {
	client client.ClientWithResponsesInterface
}

// DriverCreateBucket implements cosi.ProvisionerServer.
func (p *provisionerServer) DriverCreateBucket(context.Context, *cosi.DriverCreateBucketRequest) (*cosi.DriverCreateBucketResponse, error) {
	panic("unimplemented")
}

// DriverDeleteBucket implements cosi.ProvisionerServer.
func (p *provisionerServer) DriverDeleteBucket(context.Context, *cosi.DriverDeleteBucketRequest) (*cosi.DriverDeleteBucketResponse, error) {
	panic("unimplemented")
}

// DriverGrantBucketAccess implements cosi.ProvisionerServer.
func (p *provisionerServer) DriverGrantBucketAccess(context.Context, *cosi.DriverGrantBucketAccessRequest) (*cosi.DriverGrantBucketAccessResponse, error) {
	panic("unimplemented")
}

// DriverRevokeBucketAccess implements cosi.ProvisionerServer.
func (p *provisionerServer) DriverRevokeBucketAccess(context.Context, *cosi.DriverRevokeBucketAccessRequest) (*cosi.DriverRevokeBucketAccessResponse, error) {
	panic("unimplemented")
}
