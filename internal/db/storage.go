package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type storage struct {
	ext sqlx.ExtContext
}

type Account struct {
	AccountID string  `db:"account_id"`
	BrandID   string  `db:"brand_id"`
	Currency  string  `db:"currency"`
	Balance   float64 `db:"balance"`
}

func NewStorage(ext sqlx.ExtContext) *storage {
	return &storage{
		ext: ext,
	}
}

func (s storage) Create(ctx context.Context, account Account) error {
	_, err := sqlx.NamedExecContext(ctx, s.ext, `
			INSERT INTO account (account_id, brand_id, currency, balance) 
			VALUES (:account_id, :brand_id, :currency, :balance)
		`, account)
	return err
}

func (s storage) UpdateBalance(ctx context.Context, accountID, brandID string, amount float64) (Account, error) {
	var updatedAccount Account

	err := sqlx.GetContext(ctx, s.ext, &updatedAccount, `
			UPDATE account SET balance = balance + $1 
			WHERE account_id = $2 AND brand_id = $3 RETURNING
				account_id,
				brand_id,
				currency,
				balance
		`, amount, accountID, brandID)
	if err != nil {
		return Account{}, err
	}

	return updatedAccount, nil
}

func (s storage) GetAccount(ctx context.Context, accountID, brandID string) (Account, error) {
	var account Account

	err := sqlx.GetContext(ctx, s.ext, &account, `
			SELECT
				account_id,
				brand_id,
				currency,
				balance
			FROM account 
			WHERE account_id = $1 AND brand_id = $2
		`, accountID, brandID)
	if err != nil {
		return Account{}, err
	}

	return account, nil
}

func (s storage) Transactional(ctx context.Context, f func(tx interface{}) error) error {
	db, ok := s.ext.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("transactional: ext must be *sqlx.DB")
	}

	trx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("transactional: begin tx error: %w", err)
	}

	storage := NewStorage(trx)
	if err := f(storage); err != nil {
		_ = trx.Rollback()
		return fmt.Errorf("transactional: fn error: %w", err)
	}

	return trx.Commit()
}
