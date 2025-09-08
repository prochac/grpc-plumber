package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"time"

	pb "github.com/prochac/grpc-plumber/gen/proto/go/grpc_plumber/v1"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func main() {
	ctx := context.Background()

	serverAddr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		log.Fatalln("SERVER_ADDR env var is required")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})),
	}
	if token := os.Getenv("ACCESS_TOKEN"); token != "" {
		perRPC := oauth.TokenSource{
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: token,
			}),
		}
		opts = append(opts, grpc.WithPerRPCCredentials(perRPC))
	}

	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()
	client := pb.NewPlumberServiceClient(conn)

	for {
		resp, err := client.GetHostname(ctx, &pb.GetHostnameRequest{})
		if err != nil {
			log.Printf("failed to get hostname: %v", err)
		} else {
			log.Printf("hostname: %s", resp.Hostname)
		}
		time.Sleep(1 * time.Second)
	}
}
