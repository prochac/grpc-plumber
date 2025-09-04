package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	reflection.Register(s)
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	pb.RegisterPlumberServiceServer(s, &plumberv1.DebugServiceImplementation{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
