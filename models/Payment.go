package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

// func (m *MpesaPayment) Validate() error {
// 	if m.PhoneNumber == "" {
// 		return errors.New("phone number is required")
// 	}
// 	if m.Amount == 0 {
// 		return errors.New("amount is required")
// 	}
// 	if m.Code == "" {
// 		return errors.New("code is required")
// 	}
// 	if m.TransactionID == "" {
// 		return errors.New("transaction id is required")
// 	}
// 	if m.Status == "" {
// 		return errors.New("status is required")
// 	}
// 	return nil
// }

const (
	PaymentStatusPending   = "PENDING"
	PaymentStatusCompleted = "COMPLETED"
	PaymentStatusFailed    = "FAILED"
)

type MpesaPayment struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	BookingID      primitive.ObjectID `bson:"booking_id" json:"booking_id"`
	Amount         float64            `bson:"amount" json:"amount"`
	PaymentMethod  string             `bson:"payment_method" json:"payment_method"`
	Status         string             `bson:"status" json:"status"`
	MerchantReqID  string             `bson:"merchant_request_id,omitempty" json:"merchant_request_id,omitempty"`
	CheckoutReqID  string             `bson:"checkout_request_id,omitempty" json:"checkout_request_id,omitempty"`
	PhoneNumber    string             `bson:"phone_number,omitempty" json:"phone_number,omitempty"`
	MpesaReceiptNo string             `bson:"mpesa_receipt_no,omitempty" json:"mpesa_receipt_no,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
