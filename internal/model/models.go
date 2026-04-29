package model

type TransferRequest struct {
	IdempotencyKey string
	FromWalletID   int64
	ToWalletID     int64
	Amount         int64
}

type TransferResponse struct {
	TransferID int64
	Status     string
}