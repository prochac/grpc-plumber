# gRPC Plumber

It serves to test gRPC proxies, Ingress, Gateways, and service meshes that
provide gRPC API.

## gRPC Server

Use the OCI image `ghcr.io/prochac/grpc-plumber:latest` to run the gRPC server.

### Options

- GRPC_PORT - port to listen on (mandatory, no default)
- USE_TLS - enable TLS by setting to `1` (default: 0)

### Purpose

To test server-side timout settings. The `grpc_plumber.v1.PlumberService`
provides methods that can delay responses.

## gRPC Client

Use the OCI image `ghcr.io/prochac/grpc-plumber:latest` to run the gRPC client.
Don't forget to set the command or entrypoint to `/usr/local/bin/client`, or
just `client` should work too.

### Options

- SERVER_ADDR - gRPC server address (mandatory, no default)
- ACCESS_TOKEN - access token to send in the `authorization` metadata (optional).  
  If set, the server is expected to use TLS.

### Purpose

To test load-balancing strategy, or stickiness of sessions.