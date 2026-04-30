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

	cmd, err := r.pool.Exec(ctx,
		`INSERT INTO idempotency_records (idempotency_key, request_hash, status_code, status)
		 VALUES ($1,$2,0,'IN_PROGRESS')
		 ON CONFLICT (idempotency_key) DO NOTHING`,
		key, hash,
	)

	if err != nil {
		return false, err
	}

	// FIRST TIME INSERT → proceed
	if cmd.RowsAffected() == 1 {
		return true, nil
	}

	// ALREADY EXISTS → validate
	var storedHash string
	var status string

	err = r.pool.QueryRow(ctx,
		`SELECT request_hash, status
		 FROM idempotency_records
		 WHERE idempotency_key=$1`,
		key,
	).Scan(&storedHash, &status)

	if err != nil {
		return false, err
	}

	// CASE 1: same request → allow reuse
	if storedHash == hash {
		return false, nil
	}

	// CASE 2: different request → reject
	return false, fmt.Errorf("idempotency key reused with different request")
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
		 SET status=$1,
		     status_code=$2,
		     response_payload=$3,
		     updated_at=NOW()
		 WHERE idempotency_key=$4`,
		status, code, resp, key,
	)
	return err
}
