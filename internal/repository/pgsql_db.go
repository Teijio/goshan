package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	dsn  string
	conn *pgx.Conn
}

func NewPostgresRepository(db_dsn string) (*PostgresRepository, error) {
	conn, err := pgx.Connect(context.Background(), db_dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{
		dsn:  db_dsn,
		conn: conn,
	}, nil
}

func (pr *PostgresRepository) Check() error {
	if err := pr.conn.Ping(context.Background()); err != nil {
		pr.conn.Close(context.Background())
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

func (pr *PostgresRepository) Close() error {
	return pr.conn.Close(context.Background())
}

func (pr *PostgresRepository) Get(short string) (string, error) {
	return "", nil
}
func (pr *PostgresRepository) Save(short, original string) error {
	return nil
}
