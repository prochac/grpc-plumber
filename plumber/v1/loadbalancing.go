package v1

import (
	"context"
	"os"
	"sync"

	pb "github.com/prochac/grpc-plumber/gen/proto/go/grpc_plumber/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoadBalancingServiceServer struct {
	pb.UnimplementedLoadBalancingServiceServer

	lock          sync.RWMutex
	memoryKVStore map[string]string
}

var _ pb.LoadBalancingServiceServer = (*LoadBalancingServiceServer)(nil)

func (d *LoadBalancingServiceServer) GetHostname(context.Context, *pb.GetHostnameRequest) (*pb.GetHostnameResponse, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &pb.GetHostnameResponse{Hostname: hostname}, nil
}

func (d *LoadBalancingServiceServer) SetKey(_ context.Context, req *pb.SetKeyRequest) (*pb.SetKeyResponse, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.memoryKVStore == nil {
		d.memoryKVStore = make(map[string]string)
	}
	d.memoryKVStore[req.GetKey()] = req.GetValue()
	return &pb.SetKeyResponse{}, nil
}

func (d *LoadBalancingServiceServer) GetKey(_ context.Context, req *pb.GetKeyRequest) (*pb.GetKeyResponse, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	if d.memoryKVStore == nil {
		return &pb.GetKeyResponse{Value: ""}, nil
	}
	value, ok := d.memoryKVStore[req.GetKey()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "key %q not found", req.GetKey())
	}
	return &pb.GetKeyResponse{Value: value}, nil
}
