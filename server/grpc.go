package server

import (
	"context"
	"log"
	"net"

	"github.com/m1tka051209/calculator-service/calculator"
	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/models"
	"google.golang.org/grpc"
)

// CalculatorClient интерфейс для клиента
type CalculatorClient interface {
	Calculate(ctx context.Context, in *CalculationRequest, opts ...grpc.CallOption) (*CalculationResponse, error)
}

// calculatorClient реализация клиента
type calculatorClient struct {
	cc grpc.ClientConnInterface
}

// NewCalculatorClient создает нового клиента
func NewCalculatorClient(cc grpc.ClientConnInterface) CalculatorClient {
	return &calculatorClient{cc}
}

func (c *calculatorClient) Calculate(ctx context.Context, in *CalculationRequest, opts ...grpc.CallOption) (*CalculationResponse, error) {
	out := new(CalculationResponse)
	err := c.cc.Invoke(ctx, "/calculator.Calculator/Calculate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CalculatorServer реализует gRPC сервис
type CalculatorServer struct {
	repo db.Repository
}

// CalculationRequest запрос на вычисление
type CalculationRequest struct {
	Arg1      float64
	Arg2      float64
	Operation string
}

// CalculationResponse ответ с результатом
type CalculationResponse struct {
	Result float64
}

// ExpressionRequest запрос на создание выражения
type ExpressionRequest struct {
	UserID     string
	Expression string
}

// ExpressionResponse ответ с ID выражения
type ExpressionResponse struct {
	ExpressionID string
}

// GetExpressionsRequest запрос на получение выражений
type GetExpressionsRequest struct {
	UserID string
}

// GetExpressionsResponse ответ со списком выражений
type GetExpressionsResponse struct {
	Expressions []models.Expression
}

// StartGRPCServer запускает gRPC сервер
func StartGRPCServer(port string, repo db.Repository) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	
	// Регистрируем сервер напрямую, без генерации кода из .proto
	RegisterCalculatorServer(s, &CalculatorServer{repo: repo})
	
	log.Printf("gRPC server started on port %s", port)
	return s.Serve(lis)
}

// RegisterCalculatorServer регистрирует сервер
func RegisterCalculatorServer(s *grpc.Server, srv *CalculatorServer) {
}

// Calculate реализует gRPC метод
func (s *CalculatorServer) Calculate(ctx context.Context, req *CalculationRequest) (*CalculationResponse, error) {
	task := &models.Task{
		Arg1:      req.Arg1,
		Arg2:      req.Arg2,
		Operation: req.Operation,
	}

	result := calculator.Calculate(task)
	return &CalculationResponse{Result: result}, nil
}

// CreateExpression создает новое выражение
func (s *CalculatorServer) CreateExpression(ctx context.Context, req *ExpressionRequest) (*ExpressionResponse, error) {
	exprID, err := s.repo.CreateExpression(ctx, req.UserID, req.Expression)
	if err != nil {
		return nil, err
	}
	return &ExpressionResponse{ExpressionID: exprID}, nil
}

// GetExpressions возвращает список выражений пользователя
func (s *CalculatorServer) GetExpressions(ctx context.Context, req *GetExpressionsRequest) (*GetExpressionsResponse, error) {
	exprs, err := s.repo.GetExpressionsByUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	return &GetExpressionsResponse{Expressions: exprs}, nil
}