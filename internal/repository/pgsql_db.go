package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Teijio/goshan/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgerrcode"

)

type PostgresRepository struct {
	dsn  string
	conn *pgx.Conn
}

func NewPostgresRepository(dbDsn string) (*PostgresRepository, error) {
	conn, err := pgx.Connect(context.Background(), dbDsn)
	if err != nil {
		return nil, err
	}
	if err := runMigrations(conn); err != nil {
		return nil, err
	}
	return &PostgresRepository{
		dsn:  dbDsn,
		conn: conn,
	}, nil
}

func runMigrations(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists urls("+
		"original_url varchar unique not null, "+
		"id varchar(12) unique not null, "+
		"created_by varchar(36) not null, "+
		"correlation_id varchar"+
		");")
	if err != nil {
		return err
	}

	return nil
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

func (pr *PostgresRepository) Get(short string) (models.ShortURL, error) {
	var shortURL models.ShortURL
	err := pr.conn.QueryRow(
		context.Background(),
		"SELECT original_url, id FROM urls WHERE id=$1",
		short,
	).Scan(&shortURL.OriginalURL, &shortURL.ID)
	return shortURL, err
}
func (pr *PostgresRepository) Save(shortURL models.ShortURL) error {
	// FIXME
	// FIXME HARD
	// FIXME HARDCODE
	shortURL.CreatedByID = "jopa"
	_, err := pr.conn.Exec(
		context.Background(),
		"INSERT INTO urls (original_url, id, created_by) values ($1, $2, $3)",
		shortURL.OriginalURL,
		shortURL.ID,
		shortURL.CreatedByID,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return NewNotUniqueURLError(shortURL, err)
	}
	return err
}
