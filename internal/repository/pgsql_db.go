package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Teijio/goshan/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	dsn string
	// conn *pgx.Conn
	pool *pgxpool.Pool
}

func NewPostgresRepository(dbDsn string) (*PostgresRepository, error) {
	// conn, err := pgx.Connect(context.Background(), dbDsn)
	// if err != nil {
	// 	return nil, err
	// }
	// if err := runMigrations(conn); err != nil {
	// 	return nil, err
	// }
	// return &PostgresRepository{
	// 	dsn:  dbDsn,
	// 	conn: conn,
	// }, nil
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbDsn)
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Release()

	if err := runMigrations(conn.Conn()); err != nil {
		return nil, err
	}

	return &PostgresRepository{
		dsn:  dbDsn,
		pool: pool,
	}, err
}

func runMigrations(conn *pgx.Conn) error {
	// _, err := conn.Exec(context.Background(), "DROP TABLE urls;")
	_, err := conn.Exec(context.Background(), "create table if not exists urls("+
		"original_url varchar unique not null, "+
		"id varchar(12) unique not null, "+
		"created_by varchar(36) not null, "+
		"is_deleted timestamp, "+
		"correlation_id varchar"+
		");")
	if err != nil {
		return err
	}

	return nil
}

func (pr *PostgresRepository) Check() error {
	if err := pr.pool.Ping(context.Background()); err != nil {
		pr.pool.Close()
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

func (pr *PostgresRepository) Close() error {
	pr.pool.Close()
	return nil
}

func (pr *PostgresRepository) Get(short string) (models.ShortURL, error) {
	var shortURL models.ShortURL
	err := pr.pool.QueryRow(
		context.Background(),
		"SELECT original_url, id, is_deleted FROM urls WHERE id=$1",
		short,
	).Scan(&shortURL.OriginalURL, &shortURL.ID, &shortURL.IsDeleted)
	return shortURL, err
}
func (pr *PostgresRepository) Save(shortURL models.ShortURL) error {
	_, err := pr.pool.Exec(
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

func (pr *PostgresRepository) SaveBatch(batch []models.ShortURL) error {
	_, err := pr.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"urls"},
		[]string{"original_url", "id", "created_by", "correlation_id"},
		pgx.CopyFromSlice(len(batch), func(i int) ([]any, error) {
			return []any{batch[i].OriginalURL, batch[i].ID, batch[i].CreatedByID, batch[i].CorrelationID}, nil
		}),
	)
	return err
}

func (pr *PostgresRepository) GetUsersUrls(userID string) ([]models.ShortURL, error) {
	var URLs []models.ShortURL

	rows, err := pr.pool.Query(
		context.Background(),
		"select original_url, id, created_by from urls where created_by=$1",
		userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var entry models.ShortURL
		if err = rows.Scan(&entry); err != nil {
			return nil, err
		}
		URLs = append(URLs, entry)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return URLs, nil

}
func (pr *PostgresRepository) DeleteUrls(ctx context.Context, shortURLs []models.ShortURL) error {
	if len(shortURLs) == 0 {
		return nil
	}

	deletedAt := time.Now()
	urlsToDelete := make(map[string][]string)

	for _, url := range shortURLs {
		urlsToDelete[url.CreatedByID] = append(urlsToDelete[url.CreatedByID], url.ID)
	}

	for userID, ids := range urlsToDelete {
		tx, err := pr.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin tx for user %s: %w", userID, err)
		}

		_, err = tx.Exec(
			ctx,
			"UPDATE urls SET is_deleted = $1 WHERE created_by = $2 AND id = ANY($3)",
			deletedAt,
			userID,
			ids,
		)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to delete urls for user %s: %w", userID, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit tx for user %s: %w", userID, err)
		}
	}

	return nil
}

// func (pr *PostgresRepository) DeleteUrls(ctx context.Context, shortURLS []models.ShortURL) error {
// 	if len(shortURLS) == 0 {
// 		return nil
// 	}

// 	deletedAt := time.Now()
// 	urlsToDelete := make(map[string][]string)

// 	for _, url := range shortURLS {
// 		urlsToDelete[url.CreatedByID] = append(urlsToDelete[url.CreatedByID], url.ID)
// 	}

// 	tx, err := pr.conn.Begin(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback(ctx)
// 	for userID, ids := range urlsToDelete {
// 		_, err := tx.Exec(
// 			ctx,
// 			"UPDATE urls SET is_deleted = $1 WHERE created_by = $2 AND id = ANY($3)",
// 			deletedAt,
// 			userID,
// 			ids,
// 		)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return tx.Commit(ctx)
// }
