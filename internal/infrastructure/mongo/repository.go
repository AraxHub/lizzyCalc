package mongo

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"lizzyCalc/internal/domain"
)

// operationDoc — документ в коллекции operations (без ID — в домене ID int для совместимости с PG, при чтении оставляем 0).
type operationDoc struct {
	Number1   float64   `bson:"number1"`
	Number2   float64   `bson:"number2"`
	Operation string    `bson:"operation"`
	Result    float64   `bson:"result"`
	Message   string    `bson:"message"`
	CreatedAt time.Time `bson:"created_at"`
}

// OperationRepo реализует ports.IOperationRepository для MongoDB.
type OperationRepo struct {
	client *Client
	log    *slog.Logger
}

// NewOperationRepo возвращает репозиторий операций.
func NewOperationRepo(client *Client, log *slog.Logger) *OperationRepo {
	return &OperationRepo{client: client, log: log}
}

// SaveOperation сохраняет операцию в коллекцию.
func (r *OperationRepo) SaveOperation(ctx context.Context, op domain.Operation) error {
	doc := operationDoc{
		Number1:   op.Number1,
		Number2:   op.Number2,
		Operation: op.Operation,
		Result:    op.Result,
		Message:   op.Message,
		CreatedAt: op.Timestamp,
	}
	_, err := r.client.Coll().InsertOne(ctx, doc)
	if err != nil {
		r.log.Debug("SaveOperation failed", "error", err)
		return err
	}
	return nil
}

// GetHistory возвращает историю операций (последние сначала).
func (r *OperationRepo) GetHistory(ctx context.Context) ([]domain.Operation, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.client.Coll().Find(ctx, bson.M{}, opts)
	if err != nil {
		r.log.Debug("GetHistory failed", "error", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []operationDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	list := make([]domain.Operation, 0, len(docs))
	for _, d := range docs {
		list = append(list, domain.Operation{
			ID:        0,
			Number1:   d.Number1,
			Number2:   d.Number2,
			Operation: d.Operation,
			Result:    d.Result,
			Message:   d.Message,
			Timestamp: d.CreatedAt,
		})
	}
	return list, nil
}

// Ping проверяет доступность БД.
func (r *OperationRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}
