package models

import (
	"errors"
	"time"
)

type PaymentMethod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (p *PaymentMethod) Validate() error {
	if p.ID == "" {
		return errors.New("id is required")
	}
	if p.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type MpesaPayment struct {
	PhoneNumber   string    `json:"phone_number"`
	Amount        float64   `json:"amount"`
	Code          string    `json:"code"`
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (m *MpesaPayment) Validate() error {
	if m.PhoneNumber == "" {
		return errors.New("phone number is required")
	}
	if m.Amount == 0 {
		return errors.New("amount is required")
	}
	if m.Code == "" {
		return errors.New("code is required")
	}
	if m.TransactionID == "" {
		return errors.New("transaction id is required")
	}
	if m.Status == "" {
		return errors.New("status is required")
	}
	return nil
}
