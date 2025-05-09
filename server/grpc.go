package server

import (
	"context"
	"net"
	
	"github.com/m1tka051209/calculator-service/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CalculationRequest struct {
	Expression string
	UserId     string
}

type CalculationResponse struct {
	TaskId string
	Status string
}

type GRPCServer struct {
	repo db.Repository
}

func StartGRPCServer(port string, repo db.Repository) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	RegisterCalculatorService(s, &GRPCServer{repo: repo})
	return s.Serve(lis)
}

func (s *GRPCServer) Calculate(ctx context.Context, req *CalculationRequest) (*CalculationResponse, error) {
	taskID, err := s.repo.CreateExpression(ctx, req.UserId, req.Expression)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
	}
	
	return &CalculationResponse{
		TaskId: taskID,
		Status: "pending",
	}, nil
}

func RegisterCalculatorService(s *grpc.Server, srv *GRPCServer) {
	s.RegisterService(&_CalculatorService_serviceDesc, srv)
}

var _CalculatorService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "calculator.CalculatorService",
	HandlerType: (*GRPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Calculate",
			Handler:    _CalculatorService_Calculate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server/grpc.go",
}

func _CalculatorService_Calculate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CalculationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*GRPCServer).Calculate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/calculator.CalculatorService/Calculate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*GRPCServer).Calculate(ctx, req.(*CalculationRequest))
	}
	return interceptor(ctx, in, info, handler)
}