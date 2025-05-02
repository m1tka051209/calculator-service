package server

import (
	"context"
	"net"

	"github.com/m1tka051209/calculator-service/db"
	"google.golang.org/grpc"
)

// CalculatorServer определяет gRPC-сервис калькулятора
type CalculatorServer interface {
	Calculate(context.Context, *CalculationRequest) (*CalculationResponse, error)
}

// CalculationRequest - запрос на вычисление выражения
type CalculationRequest struct {
	Expression string
	UserId     string
}

// CalculationResponse - ответ с идентификатором задачи
type CalculationResponse struct {
	TaskId string
	Status string
}

type GRPCServer struct {
	grpc.UnimplementedServer
	repo *db.Repository
}

func NewGRPCServer(repo *db.Repository) *GRPCServer {
	return &GRPCServer{repo: repo}
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

func RegisterCalculatorServer(s *grpc.Server, srv CalculatorServer) {
	// Регистрация реализуется через рефлексию
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