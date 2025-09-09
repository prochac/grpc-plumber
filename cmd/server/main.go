package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/examples/data"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/prochac/grpc-plumber/gen/proto/go/grpc_plumber/v1"
	plumberv1 "github.com/prochac/grpc-plumber/plumber/v1"
)

func main() {
	port, ok := os.LookupEnv("GRPC_PORT")
	if !ok {
		log.Fatal("GRPC_PORT env var is required")
	}

	var opts []grpc.ServerOption
	if os.Getenv("USE_TLS") == "1" {
		cert, err := tls.LoadX509KeyPair(data.Path("x509/server_cert.pem"), data.Path("x509/server_key.pem"))
		if err != nil {
			log.Fatalf("failed to load key pair: %s", err)
		}
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	}
	s := grpc.NewServer(opts...)
	reflection.Register(s)
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	pb.RegisterTimeoutServiceServer(s, &plumberv1.TimeoutServiceServer{})
	pb.RegisterLoadBalancingServiceServer(s, &plumberv1.LoadBalancingServiceServer{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
