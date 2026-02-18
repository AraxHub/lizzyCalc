package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor логирует каждый unary RPC: метод, длительность, код/ошибка (аналог HTTP request logger).
func LoggingUnaryInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	if log == nil {
		log = slog.Default()
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		latency := time.Since(start)

		attrs := []any{"method", info.FullMethod, "latency_ms", latency.Milliseconds()}
		if err != nil {
			if st, ok := status.FromError(err); ok {
				attrs = append(attrs, "grpc_code", st.Code(), "error", st.Message())
			} else {
				attrs = append(attrs, "error", err.Error())
			}
			log.Warn("grpc request", attrs...)
			return resp, err
		}
		attrs = append(attrs, "grpc_code", codes.OK)
		log.Info("grpc request", attrs...)
		return resp, nil
	}
}
