package server

import (
	"context"
	"net"

	"github.com/m1tka051209/calculator-service/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CalculatorServer interface {
	Calculate(context.Context, *CalculationRequest) (*CalculationResponse, error)
	mustEmbedUnimplementedCalculatorServer()
}

type CalculationRequest struct {
	Expression string
	UserId     string
}

type CalculationResponse struct {
	TaskId string
	Status string
}

type UnimplementedCalculatorServer struct{}

func (UnimplementedCalculatorServer) Calculate(context.Context, *CalculationRequest) (*CalculationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Calculate not implemented")
}
func (UnimplementedCalculatorServer) mustEmbedUnimplementedCalculatorServer() {}

type GRPCServer struct {
	UnimplementedCalculatorServer
	repo db.Repository
}

func NewGRPCServer(repo db.Repository) *GRPCServer {
	return &GRPCServer{repo: repo}
}

func RegisterCalculatorServer(s *grpc.Server, srv CalculatorServer) {
	s.RegisterService(&Calculator_ServiceDesc, srv)
}

var Calculator_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "calculator.Calculator",
	HandlerType: (*CalculatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Calculate",
			Handler:    _Calculator_Calculate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "calculator.proto",
}

func _Calculator_Calculate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CalculationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CalculatorServer).Calculate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/calculator.Calculator/Calculate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CalculatorServer).Calculate(ctx, req.(*CalculationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *GRPCServer) Calculate(ctx context.Context, req *CalculationRequest) (*CalculationResponse, error) {
	exprID, err := s.repo.CreateExpression(ctx, req.UserId, req.Expression)
	if err != nil {
		return nil, err
	}

	return &CalculationResponse{
		TaskId: exprID,
		Status: "pending",
	}, nil
}

func (s *GRPCServer) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	RegisterCalculatorServer(server, s)
	return server.Serve(lis)
}