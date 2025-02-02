# Kubernetes COSI Driver for Garage

This repository implements a [COSI][cosi] driver for [Garage][garage].

## Installation

> A working installation of Garage with accessible Admin API is required.

Install COSI as documented [upstream][cosi-repo]. Customizing the installation could be done with `kustomize`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - https://github.com/kubernetes-sigs/container-object-storage-interface
```

Install the driver and configure a `Secret` to provide required Garage settings. A `kustomize` definition could look like this:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: cosi-driver-garage

resources:
  - config/overlays/default
  # For Red Hat OpenShift:
  #- config/overlays/openshift

secretGenerator:
  - name: cosi-driver-garage
    literals:
      # Garage S3 endpoint.
      - GARAGE_ENDPOINT=""
      # Garage S3 region.
      - GARAGE_REGION=""
      # Garage Admin API endpoint.
      - GARAGE_ADMIN_ENDPOINT=""
      # Garage Admin API token.
      - GARAGE_ADMIN_TOKEN=""
```

> The `kustomize` overlay for Red Hat OpenShift configures an additional rolebinding for the `anyuid` SCC.

## Usage

Configure and install `BucketClass` and `BucketAccessClass` resources:

```bash
kubectl apply -f examples/bucketclass.yaml
kubectl apply -f examples/bucketaccessclass.yaml
```

Instantiate a `BucketClaim` and `BucketAccess` resource to create a bucket and corresponding secret:

```bash
kubectl apply -f examples/bucket.yaml
```

<!-- Reference -->
[cosi]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1979-object-storage-support
[cosi-repo]: https://github.com/kubernetes-sigs/container-object-storage-interface
[garage]: https://garagehq.deuxfleurs.fr
