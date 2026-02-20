package calculator

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	calculatorv1 "github.com/AraxHub/calc-proto/gen/go/calculator/v1"
	"lizzyCalc/internal/domain"
	"lizzyCalc/internal/ports"
)

// Server реализует gRPC CalculatorService, вызывает use case калькулятора.
// В REST это был бы контроллер/хэндлер; в gRPC тип зовут Server, потому что сгенерированный интерфейс называется CalculatorServiceServer (серверная сторона RPC), и реализацию по конвенции называют так же.
type Server struct {
	calculatorv1.UnimplementedCalculatorServiceServer
	uc  ports.ICalculatorUseCase
	log *slog.Logger
}

// New создаёт gRPC-сервер калькулятора.
func New(uc ports.ICalculatorUseCase, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{uc: uc, log: log}
}

// Calculate вызывает use case и возвращает результат или gRPC-ошибку.
func (s *Server) Calculate(ctx context.Context, req *calculatorv1.CalculateRequest) (*calculatorv1.CalculateResponse, error) {
	op, err := s.uc.Calculate(ctx, req.GetNumber1(), req.GetNumber2(), req.GetOperation())
	if err != nil {
		if errors.Is(err, domain.ErrUnknownOperation) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		if err.Error() == "division by zero" {
			return nil, status.Errorf(codes.InvalidArgument, "division by zero")
		}
		s.log.Error("calculate failed", "error", err)
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if op == nil {
		return &calculatorv1.CalculateResponse{}, nil
	}
	return &calculatorv1.CalculateResponse{
		Result:  op.Result,
		Message: op.Message,
	}, nil
}

// History возвращает историю операций из use case.
func (s *Server) History(ctx context.Context, _ *calculatorv1.HistoryRequest) (*calculatorv1.HistoryResponse, error) {
	list, err := s.uc.History(ctx)
	if err != nil {
		s.log.Error("history failed", "error", err)
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	items := make([]*calculatorv1.HistoryItem, len(list))
	for i, op := range list {
		items[i] = &calculatorv1.HistoryItem{
			Id:                int32(op.ID),
			Number1:           op.Number1,
			Number2:           op.Number2,
			Operation:         op.Operation,
			Result:            op.Result,
			Message:           op.Message,
			TimestampUnixNano: op.Timestamp.UnixNano(),
		}
	}
	return &calculatorv1.HistoryResponse{Items: items}, nil
}
