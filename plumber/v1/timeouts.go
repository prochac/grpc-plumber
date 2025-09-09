package v1

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"

	pb "github.com/prochac/grpc-plumber/gen/proto/go/grpc_plumber/v1"
)

type TimeoutServiceServer struct {
	pb.UnimplementedTimeoutServiceServer
}

var _ pb.TimeoutServiceServer = (*TimeoutServiceServer)(nil)

func (d *TimeoutServiceServer) GetHostname(context.Context, *pb.GetHostnameRequest) (*pb.GetHostnameResponse, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &pb.GetHostnameResponse{Hostname: hostname}, nil
}

func (d *TimeoutServiceServer) SlowUnary(ctx context.Context, req *pb.SlowUnaryRequest) (*pb.SlowUnaryResponse, error) {
	if req.GetSleepTime() == nil {
		return &pb.SlowUnaryResponse{}, nil
	}
	if err := nonBlockingSleep(ctx, getSleepTime(req)); err != nil {
		return nil, err
	}
	return &pb.SlowUnaryResponse{}, nil
}

func (d *TimeoutServiceServer) SlowServerStream(req *pb.SlowServerStreamRequest, stream grpc.ServerStreamingServer[pb.SlowServerStreamResponse]) error {
	ctx := stream.Context()
	sleepTime := getSleepTime(req)
	for i := int32(0); i < req.GetMessageCount(); i++ {
		if err := nonBlockingSleep(ctx, sleepTime); err != nil {
			return err
		}
		if err := stream.Send(&pb.SlowServerStreamResponse{MessageIndex: i}); err != nil {
			return err
		}
	}
	return nil
}

func (d *TimeoutServiceServer) SlowClientStream(stream grpc.ClientStreamingServer[pb.SlowClientStreamRequest, pb.SlowClientStreamResponse]) error {
	ctx := stream.Context()
	var count int32
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(&pb.SlowClientStreamResponse{MessageCount: count})
		}
		if err != nil {
			return err
		}
		count++
		if err := nonBlockingSleep(ctx, getSleepTime(msg)); err != nil {
			return err
		}
	}
}

func (d *TimeoutServiceServer) SlowBiDirectionStream(stream grpc.BidiStreamingServer[pb.SlowBiDirectionStreamRequest, pb.SlowBiDirectionStreamResponse]) error {
	errG, ctx := errgroup.WithContext(stream.Context())
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		errG.Go(func() error {
			for range msg.GetMessageCount() {
				if err := nonBlockingSleep(ctx, getSleepTime(msg)); err != nil {
					return err
				}
				if err := stream.Send(&pb.SlowBiDirectionStreamResponse{MessageIndex: msg.GetMessageIndex()}); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return errG.Wait()
}

func getSleepTime(msg interface{ GetSleepTime() *durationpb.Duration }) time.Duration {
	var sleepTime time.Duration
	if msg.GetSleepTime() != nil {
		sleepTime = msg.GetSleepTime().AsDuration()
	}
	return sleepTime
}

func nonBlockingSleep(ctx context.Context, sleepTime time.Duration) error {
	if sleepTime > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepTime):
		}
	}
	return nil
}
