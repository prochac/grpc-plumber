package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"time"

	pb "github.com/prochac/grpc-lb-test/gen/proto/go/hostname"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	serverAddr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		log.Fatalln("SERVER_ADDR env var is required")
	}
	conn, err := grpc.Dial(serverAddr,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})),
	)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()
	client := pb.NewHostnameServiceClient(conn)

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
