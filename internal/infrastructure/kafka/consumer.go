package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"lizzyCalc/internal/domain"
	"lizzyCalc/internal/ports"

	"github.com/segmentio/kafka-go"
)

// Consumer — обёртка над kafka.Reader, декодирует сообщения в domain.Operation и вызывает use case.
type Consumer struct {
	r   *kafka.Reader
	uc  ports.ICalculatorUseCase
	log *slog.Logger
}

// NewConsumer создаёт консьюмера по конфигу, use case и логгеру. После использования вызови Close().
func NewConsumer(cfg *Config, uc ports.ICalculatorUseCase, log *slog.Logger) *Consumer {
	c := New(cfg).Consumer()
	c.uc = uc
	c.log = log
	return c
}

// Message — сообщение из Kafka (ключ, тело, топик, партиция, offset).
type Message = kafka.Message

// Run в цикле читает сообщения, декодирует JSON в domain.Operation, вызывает uc.HandleOperationEvent и коммитит при успехе.
// Выход по отмене ctx или при ошибке чтения.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.log.Error("kafka consumer stopped", "error", err)
			return err
		}

		var op domain.Operation
		if err := json.Unmarshal(msg.Value, &op); err != nil {
			c.log.Warn("kafka unmarshal error, skip", "error", err, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)
			_ = c.CommitMessage(ctx, msg)
			continue
		}

		if err := c.uc.HandleOperationEvent(ctx, op); err != nil {
			c.log.Warn("kafka handle error, will redeliver", "error", err, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)
			continue
		}

		if err := c.CommitMessage(ctx, msg); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.log.Error("kafka consumer stopped (commit)", "error", err)
			return err
		}
	}
}

// ReadMessage блокируется до появления следующего сообщения или отмены ctx.
func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.r.ReadMessage(ctx)
}

// FetchMessage то же, что ReadMessage, но сообщение не коммитится в consumer group до вызова CommitMessage.
func (c *Consumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return c.r.FetchMessage(ctx)
}

// CommitMessage помечает сообщение как обработанное (для consumer group).
func (c *Consumer) CommitMessage(ctx context.Context, msg kafka.Message) error {
	return c.r.CommitMessages(ctx, msg)
}

// Close закрывает консьюмера.
func (c *Consumer) Close() error {
	return c.r.Close()
}
