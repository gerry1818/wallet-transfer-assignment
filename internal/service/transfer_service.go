package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gerry1818/wallet-transfer-assignment/internal/logger"
	"github.com/gerry1818/wallet-transfer-assignment/internal/metrics"
	"github.com/gerry1818/wallet-transfer-assignment/internal/model"
	"github.com/gerry1818/wallet-transfer-assignment/internal/repository"

	"go.uber.org/zap"
)

type TransferService struct {
	repo repository.Repository
}

func NewTransferService(r repository.Repository) *TransferService {
	return &TransferService{repo: r}
}

func hash(req model.TransferRequest) string {
	b, _ := json.Marshal(req)
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}

func (s *TransferService) Transfer(ctx context.Context, req model.TransferRequest) (*model.TransferResponse, int, error) {

	logger.Log.Info("transfer start", zap.String("key", req.IdempotencyKey))

	ok, err := s.repo.InsertIdempotency(ctx, req.IdempotencyKey, hash(req))
	if err != nil {
		metrics.Failure.Inc()
		return nil, 500, err
	}

	// idempotency retry
	if !ok {
		for i := 0; i < 3; i++ {
			resp, code, _ := s.repo.GetIdempotency(ctx, req.IdempotencyKey)
			if resp != "" {
				var parsed model.TransferResponse
				_ = json.Unmarshal([]byte(resp), &parsed)
				return &parsed, code, nil
			}
			time.Sleep(50 * time.Millisecond)
		}
		return nil, 409, fmt.Errorf("request in progress")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, 500, err
	}
	defer tx.Rollback(ctx)

	// ✅ lock both wallets (always same order to avoid deadlock)
	fromBal, err := tx.GetWalletForUpdate(ctx, req.FromWalletID)
	if err != nil {
		return nil, 500, err
	}

	toBal, err := tx.GetWalletForUpdate(ctx, req.ToWalletID)
	if err != nil {
		return nil, 500, err
	}

	// ✅ validate
	if fromBal < req.Amount {
		metrics.Failure.Inc()
		logger.Log.Error("transfer failed: insufficient balance",
			zap.String("idempotency_key", req.IdempotencyKey),
			zap.Int64("from_wallet_id", req.FromWalletID),
			zap.Int64("requested_amount", req.Amount),
			zap.Int64("available_balance", fromBal),
		)
		return nil, 400, fmt.Errorf("insufficient balance")
	}

	// ✅ create transfer
	tid, err := tx.CreateTransfer(ctx,
		req.FromWalletID,
		req.ToWalletID,
		req.Amount,
		req.IdempotencyKey,
	)
	if err != nil {
		return nil, 500, err
	}

	// ✅ correct balance updates
	err = tx.UpdateWallet(ctx, req.FromWalletID, fromBal-req.Amount)
	if err != nil {
		return nil, 500, err
	}

	err = tx.UpdateWallet(ctx, req.ToWalletID, toBal+req.Amount)
	if err != nil {
		return nil, 500, err
	}

	// ✅ ledger
	_ = tx.InsertLedgerEntry(ctx, req.FromWalletID, tid, "DEBIT", req.Amount)
	_ = tx.InsertLedgerEntry(ctx, req.ToWalletID, tid, "CREDIT", req.Amount)

	// ✅ update state
	_ = tx.UpdateTransferState(ctx, tid, "PROCESSED")

	// ✅ commit
	if err := tx.Commit(ctx); err != nil {
		return nil, 500, err
	}

	resp := model.TransferResponse{
		TransferID: tid,
		Status:     "PROCESSED",
	}

	b, _ := json.Marshal(resp)

	// ✅ store idempotency OUTSIDE tx
	_ = s.repo.SaveIdempotency(ctx, req.IdempotencyKey, string(b), 200)

	metrics.Success.Inc()

	logger.Log.Info("transfer success", zap.Int64("transfer_id", tid))

	return &resp, 200, nil
}
