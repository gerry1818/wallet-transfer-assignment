package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type txImpl struct {
	tx pgx.Tx
}

func (t *txImpl) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *txImpl) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *txImpl) GetBalance(ctx context.Context, walletID int64) (int64, error) {
	var balance int64
	err := t.tx.QueryRow(ctx,
		`SELECT balance FROM wallets WHERE id=$1`,
		walletID,
	).Scan(&balance)

	return balance, err
}

func (t *txImpl) GetWalletForUpdate(ctx context.Context, id int64) (int64, error) {
	var balance int64
	err := t.tx.QueryRow(ctx,
		`SELECT balance FROM wallets WHERE id=$1 FOR UPDATE`,
		id,
	).Scan(&balance)

	return balance, err
}

func (t *txImpl) UpdateWallet(ctx context.Context, id int64, balance int64) error {
	_, err := t.tx.Exec(ctx,
		`UPDATE wallets SET balance=$1 WHERE id=$2`,
		balance, id,
	)
	return err
}

func (t *txImpl) CreateTransfer(ctx context.Context, from, to, amount int64, key string) (int64, error) {
	var id int64
	err := t.tx.QueryRow(ctx,
		`INSERT INTO transfers (from_wallet_id, to_wallet_id, amount, state, idempotency_key)
		 VALUES ($1,$2,$3,'PENDING',$4) RETURNING id`,
		from, to, amount, key,
	).Scan(&id)

	return id, err
}

func (t *txImpl) UpdateTransferState(ctx context.Context, transferID int64, state string) error {
	_, err := t.tx.Exec(ctx,
		`UPDATE transfers SET state=$1 WHERE id=$2`,
		state, transferID,
	)
	return err
}

func (t *txImpl) InsertLedgerEntry(ctx context.Context, walletID, transferID int64, typ string, amount int64) error {
	_, err := t.tx.Exec(ctx,
		`INSERT INTO ledger_entries (wallet_id, transfer_id, type, amount)
		 VALUES ($1,$2,$3,$4)`,
		walletID, transferID, typ, amount,
	)
	return err
}
