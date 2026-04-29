package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gerry1818/wallet-transfer-assignment/internal/model"
	"github.com/gerry1818/wallet-transfer-assignment/internal/service"
)

type Handler struct {
	svc *service.TransferService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewHandler(s *service.TransferService) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid request body",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp, code, err := h.svc.Transfer(ctx, req)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(resp)
}
