package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"

	pb "github.com/prochac/grpc-plumber/gen/proto/go/grpc_plumber/v1"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	ctx := context.Background()

	serverAddr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		log.Fatalln("SERVER_ADDR env var is required")
	}

	var opts []grpc.DialOption
	if os.Getenv("USE_TLS") == "1" {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			})),
		}
	} else {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}

	if token := os.Getenv("ACCESS_TOKEN"); token != "" {
		perRPC := oauth.TokenSource{
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: token,
			}),
		}
		opts = append(opts, grpc.WithPerRPCCredentials(perRPC))
	}
	if err := getHostnames(ctx, serverAddr, opts...); err != nil {
		log.Fatalf("failed to get hostnames: %v", err)
	}
	if sessionHeader := os.Getenv("SESSION_HEADER"); sessionHeader != "" {
		if err := testStickySession(ctx, sessionHeader, serverAddr, opts...); err != nil {
			log.Fatalf("failed to get hostnames: %v", err)
		}
	}
}

func getHostnames(ctx context.Context, serverAddr string, opts ...grpc.DialOption) error {
	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}
	defer conn.Close()
	client := pb.NewLoadBalancingServiceClient(conn)

	// Test the load balancing by making multiple requests and printing the hostname of the server handling each request.
	for range 10 {
		resp, err := client.GetHostname(ctx, &pb.GetHostnameRequest{})
		if err != nil {
			return fmt.Errorf("failed to get hostname: %w", err)
		}
		log.Printf("hostname: %s", resp.Hostname)
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}

func testStickySession(ctx context.Context, sessionHeader string, serverAddr string, opts ...grpc.DialOption) error {
	readConn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}
	defer readConn.Close()
	readClient := pb.NewLoadBalancingServiceClient(readConn)

	writeConn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}
	defer writeConn.Close()
	writeClient := pb.NewLoadBalancingServiceClient(writeConn)

	for range 10 {
		sessionID := randomString(32)
		ctxWithSession := metadata.AppendToOutgoingContext(ctx, sessionHeader, sessionID)
		randKey, randValue := randomString(32), randomString(32)
		// Write a value
		_, err := writeClient.SetKey(ctxWithSession, &pb.SetKeyRequest{
			Key:   randKey,
			Value: randValue,
		})
		if err != nil {
			return fmt.Errorf("failed to set key: %w", err)
		}
		// Read the value back
		resp, err := readClient.GetKey(ctxWithSession, &pb.GetKeyRequest{
			Key: randKey,
		})
		if status.Code(err) == 5 {
			log.Printf("Sticky session failed: key %q not found", randKey)
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to get key: %w", err)
		}
		if resp.Value != randValue {
			log.Printf("Sticky session failed: expected %q, got %q", randValue, resp.Value)
		}
		log.Printf("Sticky session succeeded: key %q has value %q", randKey, resp.Value)
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}
