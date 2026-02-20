package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Producer — обёртка над kafka.Writer для отправки сообщений в топик.
type Producer struct {
	w *kafka.Writer
}

// NewProducer создаёт продюсера по конфигу. После использования вызови Close().
func NewProducer(cfg *Config) *Producer {
	return New(cfg).Producer()
}

// Send отправляет одно сообщение (key и value — произвольные байты).
func (p *Producer) Send(ctx context.Context, key, value []byte) error {
	return p.w.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}


// Close закрывает продюсера.
func (p *Producer) Close() error {
	return p.w.Close()
}
