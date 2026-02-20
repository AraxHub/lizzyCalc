package kafka

import (
	"strings"

	"github.com/segmentio/kafka-go"
)

// Config — настройки Kafka. Переменные: CALCULATOR_KAFKA_BROKERS, CALCULATOR_KAFKA_TOPIC, CALCULATOR_KAFKA_GROUP_ID.
type Config struct {
	Brokers string `envconfig:"BROKERS" default:"localhost:9092"` // через запятую, если несколько
	Topic   string `envconfig:"TOPIC" default:"lizzycalc"`
	GroupID string `envconfig:"GROUP_ID" default:"lizzycalc-app"` // для consumer group
}

// brokersSlice возвращает список брокеров из строки (через запятую).
func (c *Config) brokersSlice() []string {
	if c == nil || c.Brokers == "" {
		return []string{"localhost:9092"}
	}
	parts := strings.Split(c.Brokers, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// Client — конфиг и фабрики продюсера/консьюмера. Подключение к брокеру при создании Writer/Reader.
type Client struct {
	cfg *Config
}

// New создаёт клиент по конфигу. Само подключение к Kafka — при первом вызове Producer() или Consumer().
func New(cfg *Config) *Client {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Client{cfg: cfg}
}

// Producer создаёт продюсера для отправки сообщений в топик. После использования вызови Close().
func (c *Client) Producer() *Producer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(c.cfg.brokersSlice()...),
		Topic:    c.cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{w: w}
}

// Consumer создаёт консьюмера для чтения из топика (consumer group). После использования вызови Close().
func (c *Client) Consumer() *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.cfg.brokersSlice(),
		Topic:   c.cfg.Topic,
		GroupID: c.cfg.GroupID,
	})
	return &Consumer{r: r}
}
