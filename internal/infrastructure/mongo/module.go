package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Config — настройки подключения к MongoDB. Переменные: CALCULATOR_MONGO_*.
type Config struct {
	URI        string `envconfig:"URI" default:"mongodb://localhost:27017"`
	Database   string `envconfig:"DATABASE" default:"lizzycalc"`
	Collection string `envconfig:"COLLECTION" default:"operations"`
}

// Client — обёртка над mongo.Client.
type Client struct {
	*mongo.Client
	cfg Config
}

// New подключается к MongoDB по конфигу.
func New(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	return &Client{Client: client, cfg: *cfg}, nil
}

// DB возвращает базу по конфигу.
func (c *Client) DB() *mongo.Database {
	return c.Database(c.cfg.Database)
}

// Coll возвращает коллекцию операций.
func (c *Client) Coll() *mongo.Collection {
	return c.DB().Collection(c.cfg.Collection)
}
