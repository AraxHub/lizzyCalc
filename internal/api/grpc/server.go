package grpc

import (
	"context"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	calculatorv1 "github.com/AraxHub/calc-proto/gen/go/calculator/v1"
	"lizzyCalc/internal/api/grpc/calculator"
	"lizzyCalc/internal/api/grpc/interceptors"
	"lizzyCalc/internal/ports"
)

// Server — gRPC-сервер: регистрирует сервисы и слушает порт.
type Server struct {
	grpc *grpc.Server
	uc   ports.ICalculatorUseCase
	addr string
}

// NewServer создаёт gRPC-сервер и регистрирует CalculatorService. Логирующий интерцептор пишет метод, latency_ms и grpc_code (аналог HTTP middleware).
func NewServer(addr string, uc ports.ICalculatorUseCase, log *slog.Logger) *Server {
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors.LoggingUnaryInterceptor(log)))
	calculatorv1.RegisterCalculatorServiceServer(s, calculator.New(uc, log))
	return &Server{grpc: s, uc: uc, addr: addr}
}

// Start слушает addr и принимает соединения (блокируется). Остановка через Stop().
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.grpc.Serve(lis)
}

// Stop останавливает сервер (graceful).
func (s *Server) Stop(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.grpc.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.grpc.Stop()
		return ctx.Err()
	}
}
