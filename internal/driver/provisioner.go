package driver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/mpreu/cosi-driver-garage/internal/client"
	"github.com/mpreu/cosi-driver-garage/internal/config"
)

// Interface assert.
var _ cosi.ProvisionerServer = &provisionerServer{}

// provisionerServer implements cosi.ProvisionerServer.
type provisionerServer struct {
	cosi.UnimplementedProvisionerServer
	client client.ClientWithResponsesInterface
	config *config.Garage
	logger *slog.Logger
}

// DriverCreateBucket implements cosi.ProvisionerServer.
//
// Notes from specification:
// This call is made to create the bucket in the backend.
// This call is idempotent
//  1. If a bucket that matches both name and parameters already exists, then OK (success) must be returned.
//  2. If a bucket by same name, but different parameters is provided, then the appropriate error code ALREADY_EXISTS must be returned.
func (p *provisionerServer) DriverCreateBucket(ctx context.Context, r *cosi.DriverCreateBucketRequest) (*cosi.DriverCreateBucketResponse, error) {
	logger := p.logger.With("req", r)
	logger.Info("DriverCreateBucket request")

	name := r.GetName()

	// Check if bucket already exists.
	existingID, err := p.hasBucket(ctx, name)
	if err != nil {
		logger.Error("Failed to check for existing bucket", "error", err)
		return nil, status.Error(codes.Internal, "failed to check for existing bucket")
	}

	// Since no parameters apart from the name are accepted, return without error.
	if existingID != nil {
		return &cosi.DriverCreateBucketResponse{
			BucketId:   *existingID,
			BucketInfo: p.protocol(),
		}, nil
	}

	// Otherwise, create a new bucket.
	req := client.CreateBucketJSONRequestBody{
		GlobalAlias: &name,
	}

	resp, err := p.client.CreateBucketWithResponse(ctx, req)
	if err != nil {
		logger.Error("Failed to create bucket", "error", err)
		return nil, status.Error(codes.Internal, "failed to create bucket")
	}

	if code := resp.StatusCode(); code != http.StatusOK {
		logger.Error("Failed to create bucket with unexpected HTTP status code",
			"httpStatusExpected", http.StatusOK,
			"httpStatusGot", code)

		return nil, status.Error(codes.Internal, "failed to create bucket")
	}

	return &cosi.DriverCreateBucketResponse{
		BucketId:   *resp.JSON200.Id,
		BucketInfo: p.protocol(),
	}, nil
}

// DriverDeleteBucket implements cosi.ProvisionerServer.
//
// Notes from specification:
// This call is made to delete the bucket in the backend.
// If the bucket has already been deleted, then no error should be returned.
func (p *provisionerServer) DriverDeleteBucket(ctx context.Context, r *cosi.DriverDeleteBucketRequest) (*cosi.DriverDeleteBucketResponse, error) {
	logger := p.logger.With("req", r)
	logger.Info("DriverDeleteBucket request")

	resp, err := p.client.DeleteBucketWithResponse(ctx, &client.DeleteBucketParams{Id: r.GetBucketId()})
	if err != nil {
		logger.Error("Failed to delete bucket", "error", err)
		return nil, status.Error(codes.Internal, "failed to delete bucket")
	}

	// If a bucket is not found, this is a no-op.
	code := resp.StatusCode()
	if code != http.StatusNoContent && code != http.StatusNotFound {
		logger.Error("Failed to delete bucket with unexpected HTTP status code",
			"httpStatusExpected", http.StatusNoContent,
			"httpStatusGot", code)

		return nil, status.Error(codes.Internal, "failed to delete bucket")
	}

	return &cosi.DriverDeleteBucketResponse{}, nil
}

// DriverGrantBucketAccess implements cosi.ProvisionerServer.
func (p *provisionerServer) DriverGrantBucketAccess(ctx context.Context, r *cosi.DriverGrantBucketAccessRequest) (*cosi.DriverGrantBucketAccessResponse, error) {
	logger := p.logger.With("req", r)
	logger.Info("DriverGrantBucketAccess request")

	// Check for authentication type.
	if r.AuthenticationType == cosi.AuthenticationType_IAM {
		logger.Error("Authentication type IAM not implemented")
		return nil, status.Error(codes.Unimplemented, "authentication type IAM not implemented")
	}

	// Create new API key.
	// TODO: Tokens with same name are possible. Guard against it?
	accountName := r.GetName()
	keyResp, err := p.client.AddKeyWithResponse(ctx, client.AddKeyJSONRequestBody{Name: &accountName})
	if err != nil {
		logger.Error("Failed to create key", "error", err)
		return nil, status.Error(codes.Internal, "failed to create key")
	}

	if code := keyResp.StatusCode(); code != http.StatusOK {
		logger.Error("Failed to create key with unexpected HTTP status code",
			"httpStatusExpected", http.StatusOK,
			"httpStatusGot", code)

		return nil, status.Error(codes.Internal, "failed to create key")
	}

	s3AccessKeyID := *keyResp.JSON200.AccessKeyId
	s3AccessKey := *keyResp.JSON200.SecretAccessKey

	permissions, err := permissions(r.Parameters)
	if err != nil {
		logger.Error("Failed to parse BucketAccessClass parameters", "error", err)
		return nil, status.Error(codes.InvalidArgument, "failed to parse BucketAccessClass parameters")
	}

	// Assign key to bucket.
	req := client.AllowBucketKeyJSONRequestBody{
		AccessKeyId: s3AccessKeyID,
		BucketId:    r.BucketId,
		Permissions: struct {
			Owner bool "json:\"owner\""
			Read  bool "json:\"read\""
			Write bool "json:\"write\""
		}{
			Owner: permissions.owner,
			Read:  permissions.read,
			Write: permissions.write,
		},
	}

	allowResp, err := p.client.AllowBucketKeyWithResponse(ctx, req)
	if err != nil {
		logger.Error("Failed to assign key to bucket", "error", err)
		return nil, status.Error(codes.Internal, "failed to assign key to bucket")
	}

	if code := allowResp.StatusCode(); code != http.StatusOK {
		logger.Error("Failed to assign key to bucket with unexpected HTTP status code",
			"httpStatusExpected", http.StatusOK,
			"httpStatusGot", code)

		return nil, status.Error(codes.Internal, "failed to assign key to bucket")
	}

	return &cosi.DriverGrantBucketAccessResponse{
		AccountId: s3AccessKeyID,
		Credentials: map[string]*cosi.CredentialDetails{
			"s3": p.s3Credentials(s3AccessKeyID, s3AccessKey),
		},
	}, nil
}

// DriverRevokeBucketAccess implements cosi.ProvisionerServer.
func (p *provisionerServer) DriverRevokeBucketAccess(ctx context.Context, r *cosi.DriverRevokeBucketAccessRequest) (*cosi.DriverRevokeBucketAccessResponse, error) {
	logger := p.logger.With("req", r)
	logger.Info("DriverRevokeBucketAccess request")

	resp, err := p.client.DeleteKeyWithResponse(ctx, &client.DeleteKeyParams{Id: r.AccountId})
	if err != nil {
		logger.Error("Failed to delete key", "error", err)
		return nil, status.Error(codes.Internal, "failed to delete key")
	}

	if code := resp.StatusCode(); code != http.StatusNoContent {
		logger.Error("Failed to delete key with unexpected HTTP status code",
			"httpStatusExpected", http.StatusNoContent,
			"httpStatusGot", code)

		return nil, status.Error(codes.Internal, "failed to delete key")
	}

	return &cosi.DriverRevokeBucketAccessResponse{}, nil
}

// hasBucket checks if a bucket already exists and returns its ID.
func (p *provisionerServer) hasBucket(ctx context.Context, name string) (*string, error) {
	list, err := p.client.ListBucketsWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if list.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("error listing buckets, HTTP status code %d", list.StatusCode())
	}

	var id *string
	for _, l := range *list.JSON200 {
		if l.GlobalAliases != nil {
			if slices.Contains(*l.GlobalAliases, name) {
				id = &l.Id
				break
			}
		}
	}

	return id, nil
}

// protocol returns details of supported object bucket protocols.
func (p *provisionerServer) protocol() *cosi.Protocol {
	return &cosi.Protocol{
		Type: &cosi.Protocol_S3{
			S3: &cosi.S3{
				Region:           p.config.Region,
				SignatureVersion: cosi.S3SignatureVersion_S3V4,
			},
		},
	}
}

// s3Credentials returns the credentials for the S3 protocol.
func (p *provisionerServer) s3Credentials(keyID, key string) *cosi.CredentialDetails {
	return &cosi.CredentialDetails{
		Secrets: map[string]string{
			"endpoint":        p.config.Endpoint,
			"region":          p.config.Region,
			"accessKeyID":     keyID,
			"accessSecretKey": key,
		},
	}
}

// accessPermissions represents possible bucket access key permissions.
type accessPermissions struct {
	owner bool
	read  bool
	write bool
}

// permissions parses the bucket access permissions from BucketAccessClass parameters.
func permissions(params map[string]string) (*accessPermissions, error) {
	p := &accessPermissions{
		owner: false,
		read:  true,
		write: true,
	}

	if params == nil {
		return p, nil
	}

	if v, ok := params["owner"]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}

		p.owner = b
	}

	if v, ok := params["read"]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}

		p.read = b
	}

	if v, ok := params["write"]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}

		p.write = b
	}

	return p, nil
}
