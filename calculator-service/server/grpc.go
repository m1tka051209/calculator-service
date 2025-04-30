package server

import (
    "context"
    // "log"
    // "net"
    "github.com/m1tka051209/calculator-service/proto"
    "github.com/m1tka051209/calculator-service/task_manager"
    // "google.golang.org/grpc"
)

type grpcServer struct {
    proto.UnimplementedArithmeticServiceServer
    tm *task_manager.TaskManager
}

func (s *grpcServer) SubmitResult(ctx context.Context, res *proto.Result) (*proto.Empty, error) {
    if err := s.tm.SaveTaskResult(res.TaskId, res.Value); err != nil {
        return nil, err
    }
    return &proto.Empty{}, nil
}