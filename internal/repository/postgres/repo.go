package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gerry1818/wallet-transfer-assignment/internal/repository"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(p *pgxpool.Pool) *Repo {
	return &Repo{pool: p}
}

func (r *Repo) InsertIdempotency(ctx context.Context, key, hash string) (bool, error) {
	cmd, err := r.pool.Exec(ctx,
		`INSERT INTO idempotency_records (idempotency_key, request_hash, status_code)
		 VALUES ($1,$2,0)
		 ON CONFLICT DO NOTHING`,
		key, hash)

	return cmd.RowsAffected() == 1, err
}

func (r *Repo) GetIdempotency(ctx context.Context, key string) (string, int, error) {
	var resp string
	var code int

	err := r.pool.QueryRow(ctx,
		`SELECT response_payload, status_code FROM idempotency_records WHERE idempotency_key=$1`,
		key).Scan(&resp, &code)

	return resp, code, err
}

func (r *Repo) BeginTx(ctx context.Context) (repository.Tx, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &txImpl{
		tx: tx,
	}, nil
}

func (r *Repo) SaveIdempotency(ctx context.Context, key, resp string, status int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE idempotency_records
		 SET response_payload=$1, status_code=$2
		 WHERE idempotency_key=$3`,
		resp, status, key,
	)
	return err
}
