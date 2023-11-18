package contract

import "fmt"

type CreateRequest struct {
	AccountID string  `json:"account_id"`
	BrandID   string  `json:"brand_id"`
	Balance   float64 `json:"balance"`
	Currency  string  `json:"currency"`
}

type UpdateRequest struct {
	AccountID string  `json:"account_id"`
	BrandID   string  `json:"brand_id"`
	Amount    float64 `json:"amount"`
}

type GetBalanceRequest struct {
	AccountID string `json:"account_id"`
	BrandID   string `json:"brand_id"`
}

type BalanceResponse struct {
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

func (r CreateRequest) Validate() error {
	if r.BrandID == "" {
		return fmt.Errorf("brand_id shoud not be empty")
	}

	if r.AccountID == "" {
		return fmt.Errorf("account_id shoud not be empty")
	}

	if r.Currency == "" {
		return fmt.Errorf("currency shoud not be empty")
	}

	if r.Balance < 0 {
		return fmt.Errorf("balance should be gte 0")
	}

	return nil
}

func (r UpdateRequest) Validate() error {
	if r.BrandID == "" {
		return fmt.Errorf("brand_id shoud not be empty")
	}

	if r.AccountID == "" {
		return fmt.Errorf("account_id shoud not be empty")
	}

	if r.Amount < 0 {
		return fmt.Errorf("amount should be gte 0")
	}

	return nil
}

func (r GetBalanceRequest) Validate() error {
	if r.BrandID == "" {
		return fmt.Errorf("brand_id shoud not be empty")
	}

	if r.AccountID == "" {
		return fmt.Errorf("account_id shoud not be empty")
	}

	return nil
}
