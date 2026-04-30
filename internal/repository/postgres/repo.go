package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gerry1818/wallet-transfer-assignment/internal/repository"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(p *pgxpool.Pool) *Repo {
	return &Repo{pool: p}
}

func (r *Repo) ClaimIdempotency(ctx context.Context, key, hash string) (bool, error) {
	var storedHash string

	err := r.pool.QueryRow(ctx,
		`INSERT INTO idempotency_records (idempotency_key, request_hash, status_code)
		 VALUES ($1,$2,0)
		 ON CONFLICT (idempotency_key)
		 DO UPDATE SET idempotency_key = EXCLUDED.idempotency_key
		 RETURNING request_hash`,
		key, hash,
	).Scan(&storedHash)

	if err != nil {
		// real error
		return false, err
	}

	// If returned hash is empty → new insert
	if storedHash == "" {
		return true, nil
	}

	// If hash mismatch → invalid reuse
	if storedHash != hash {
		return false, fmt.Errorf("idempotency key reused with different request")
	}

	return false, nil // same request (duplicate)
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

func (r *Repo) UpdateIdempotency(ctx context.Context, key, status, resp string, code int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE idempotency_records
		 SET status_code=$1,
		     response_payload=$2
		 WHERE idempotency_key=$3`,
		code, resp, key,
	)
	return err
}
