package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
}

func NewPostgreSQLDB(dsn string) (*DB, error) {
	config, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, err
	}

	// Connection pool settings
	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.Pool.Begin(ctx)
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}
