package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/scorum/account-svc/internal/db"
	"github.com/scorum/account-svc/internal/handler/contract"
	"github.com/sirupsen/logrus"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Storage interface {
	Create(ctx context.Context, account db.Account) error
	UpdateBalance(ctx context.Context, accountID, brandID string, amount float64) (db.Account, error)
	GetAccount(ctx context.Context, accountID, brandID string) (db.Account, error)
	Transactional(ctx context.Context, f func(tx interface{}) error) error
}

type Handler struct {
	storage Storage
	log     *logrus.Entry
}

func New(storage Storage) *Handler {
	return &Handler{
		storage: storage,
		log:     logrus.WithField("handler", "api"),
	}
}

func (h Handler) transact(ctx context.Context, do func(Storage) error) error {
	return h.storage.Transactional(ctx, func(txI interface{}) error {
		if tx, ok := txI.(Storage); ok {
			return do(tx)
		}
		return errors.New("unsupported store type")
	})
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req contract.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internalServerError(w, err, "failed to decode", h.log)
		return
	}

	if err := req.Validate(); err != nil {
		warnInvalidRequest(w, http.StatusBadRequest, err.Error(), h.log)
		return
	}

	if err := h.storage.Create(r.Context(), db.Account{
		AccountID: req.AccountID,
		BrandID:   req.BrandID,
		Balance:   req.Balance,
		Currency:  req.Currency,
	}); err != nil {
		internalServerError(w, err, "failed to create player", h.log)
		return
	}

	render.JSON(w, r, contract.BalanceResponse{
		Balance:  req.Balance,
		Currency: req.Currency,
	})
}

func (h Handler) Credit(w http.ResponseWriter, r *http.Request) {
	var req contract.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internalServerError(w, err, "failed to decode", h.log)
		return
	}

	if err := req.Validate(); err != nil {
		warnInvalidRequest(w, http.StatusBadRequest, err.Error(), h.log)
		return
	}

	account, err := h.storage.UpdateBalance(r.Context(), req.AccountID, req.BrandID, req.Amount)
	if err != nil {
		internalServerError(w, err, "failed to credit account", h.log)
		return
	}

	render.JSON(w, r, contract.BalanceResponse{
		Balance:  account.Balance,
		Currency: account.Currency,
	})
}

func (h Handler) Debit(w http.ResponseWriter, r *http.Request) {
	var req contract.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internalServerError(w, err, "failed to decode", h.log)
		return
	}

	if err := req.Validate(); err != nil {
		warnInvalidRequest(w, http.StatusBadRequest, err.Error(), h.log.WithError(err))
		return
	}

	var account db.Account
	err := h.transact(r.Context(), func(storage Storage) error {
		var err error
		account, err = storage.GetAccount(r.Context(), req.AccountID, req.BrandID)
		if err != nil {
			return fmt.Errorf("get account: %w", err)
		}

		if account.Balance-req.Amount < 0 {
			return ErrInsufficientFunds
		}

		account, err = storage.UpdateBalance(r.Context(), req.AccountID, req.BrandID, -req.Amount)
		if err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		return nil
	})

	if err != nil {
		internalServerError(w, err, "failed to debit account", h.log)
		return
	}

	render.JSON(w, r, contract.BalanceResponse{
		Balance:  account.Balance,
		Currency: account.Currency,
	})
}

func (h Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	var req contract.GetBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internalServerError(w, err, "failed to decode", h.log)
		return
	}

	if err := req.Validate(); err != nil {
		warnInvalidRequest(w, http.StatusBadRequest, err.Error(), h.log)
		return
	}

	account, err := h.storage.GetAccount(r.Context(), req.AccountID, req.BrandID)
	if err != nil {
		internalServerError(w, err, "failed to get account", h.log)
		return
	}

	render.JSON(w, r, contract.BalanceResponse{
		Balance:  account.Balance,
		Currency: account.Currency,
	})
}

func internalServerError(w http.ResponseWriter, err error, msg string, log *logrus.Entry) {
	log.WithError(err).Error(msg)
	w.WriteHeader(http.StatusInternalServerError)
}

func warnInvalidRequest(w http.ResponseWriter, statusCode int, msg string, log *logrus.Entry) {
	log.Warn(msg)
	w.WriteHeader(statusCode)
}
