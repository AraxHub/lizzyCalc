package ports

//go:generate mockgen -source=broker.go -destination=../mocks/broker_mock.go -package=mocks

import "context"

// IProducer — контракт отправки сообщений в брокер (например Kafka). Топик задаётся при создании реализации (конфиг).
// Use case после расчёта вызывает Send; консьюмер живёт в инфраструктуре.
type IProducer interface {
	Send(ctx context.Context, key, value []byte) error
}
