##
## Build stage.
##
FROM docker.io/library/golang:1.23 as builder

# Set the working directory.
WORKDIR /workspace

# Prepare dir so it can be copied over to runtime layer.
RUN mkdir -p /var/lib/cosi

# Copy the Go Modules manifests.
COPY go.mod go.mod
COPY go.sum go.sum

# Cache dependencies.
RUN go mod download \
    && go install github.com/go-delve/delve/cmd/dlv@latest

# Copy the source code.
COPY cmd/ cmd/
COPY internal/ internal/
COPY tools/ tools/

# Build.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -trimpath -a -o bin/cosi-driver-garage cmd/main.go

#
# Debug image.
#
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.3 as debug

ENV LC_ALL=C.UTF-8
WORKDIR /

COPY --from=builder --chown=65532:0 /workspace/bin/cosi-driver-garage /usr/bin/cosi-driver-garage
COPY --from=builder --chown=65532:0 /go/bin/dlv /usr/bin/dlv

# Args for labels.
ARG VERSION=debug

# Add labels.
LABEL org.opencontainers.image.title="cosi-driver-garage"
LABEL org.opencontainers.image.description="COSI Driver for Garage"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.license="MIT"
LABEL org.opencontainers.image.source="https://github.com/mpreu/cosi-driver-garage"
LABEL org.opencontainers.image.documentation="https://github.com/mpreu/cosi-driver-garage"
LABEL org.opencontainers.image.base.name="registry.access.redhat.com/ubi9/ubi-minimal:9.3"

EXPOSE 40000
ENTRYPOINT ["/usr/bin/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--log", "exec", "/usr/bin/cosi-driver-garage"]

##
## Runtime image.
##
FROM gcr.io/distroless/static:nonroot AS runtime

# Copy the executable.
COPY --from=builder --chown=65532:0 /workspace/bin/cosi-driver-garage /usr/bin/cosi-driver-garage

# Copy the volume directory with correct permissions, so driver can bind a socket there.
COPY --from=builder --chown=65532:0 /var/lib/cosi /var/lib/cosi

# Set volume mount point for app socket.
VOLUME [ "/var/lib/cosi" ]

# Set the final UID:GID to non-root user.
USER 65532:65532

# Args for labels.
ARG VERSION=dev

# Add labels.
LABEL org.opencontainers.image.title="cosi-driver-garage"
LABEL org.opencontainers.image.description="COSI Driver for Garage"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.license="MIT"
LABEL org.opencontainers.image.source="https://github.com/mpreu/cosi-driver-garage"
LABEL org.opencontainers.image.documentation="https://github.com/mpreu/cosi-driver-garage"
LABEL org.opencontainers.image.base.name="gcr.io/distroless/static:latest"

# Set the entrypoint.
ENTRYPOINT [ "/usr/bin/cosi-driver-garage" ]

