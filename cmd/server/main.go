package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/prochac/grpc-lb-test/gen/proto/go/hostname"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedHostnameServiceServer
}

func (s *server) GetHostname(_ context.Context, _ *pb.GetHostnameRequest) (*pb.GetHostnameResponse, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &pb.GetHostnameResponse{Hostname: hostname}, nil
}

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
	pb.RegisterHostnameServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
