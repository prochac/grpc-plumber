package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/prochac/grpc-plumber/gen/proto/go/grpc_plumber/v1"
)

func main() {
	serverAddr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		log.Fatalln("SERVER_ADDR env var is required")
	}
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})),
	)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()
	client := pb.NewPlumberServiceClient(conn)

	for {
		resp, err := client.GetHostname(context.Background(), &pb.GetHostnameRequest{})
		if err != nil {
			log.Printf("failed to get hostname: %v", err)
		} else {
			log.Printf("hostname: %s", resp.Hostname)
		}
		time.Sleep(1 * time.Second)
	}
}
