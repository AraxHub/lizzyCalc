package pg

import (
	"context"
)

const createOperationsTable = `
CREATE TABLE IF NOT EXISTS operations (
	id         SERIAL PRIMARY KEY,
	number1    DOUBLE PRECISION NOT NULL,
	number2    DOUBLE PRECISION NOT NULL,
	operation  VARCHAR(10) NOT NULL,
	result     DOUBLE PRECISION NOT NULL,
	message    TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// Migrate создаёт таблицу operations, если её ещё нет.
func Migrate(ctx context.Context, db *DB) error {
	_, err := db.ExecContext(ctx, createOperationsTable)
	return err
}
