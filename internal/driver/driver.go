package driver

import cosi "sigs.k8s.io/container-object-storage-interface-spec"

// New returns implementations for the COSI.IdentityServer and
// cosi.ProvisionerServer interfaces.
func New() (cosi.IdentityServer, cosi.ProvisionerServer) {
	is := &identityServer{}
	ps := &provisionerServer{}

	return is, ps
}
