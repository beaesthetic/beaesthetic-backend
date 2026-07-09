package fidelity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Treatment string

const TreatmentSolarium Treatment = "SOLARIUM"

type Voucher struct {
	ID        string    `json:"id"`
	IssuedAt  time.Time `json:"issuedAt"`
	IsUsed    bool      `json:"isUsed"`
	Treatment Treatment `json:"treatment"`
}

type Card struct {
	ID                string
	CustomerID        string
	SolariumPurchases int
	Vouchers          []Voucher
}

func NewCard(customerID string) Card { return Card{ID: uuid.NewString(), CustomerID: customerID} }

func (card Card) RegisterPurchase(treatment Treatment) (Card, *Voucher, error) {
	if treatment == "" {
		treatment = TreatmentSolarium
	}
	if treatment != TreatmentSolarium {
		return Card{}, nil, fmt.Errorf("unsupported treatment %s", treatment)
	}
	card.SolariumPurchases++
	if card.SolariumPurchases >= 10 {
		voucher := Voucher{ID: uuid.NewString(), IssuedAt: time.Now().UTC(), IsUsed: false, Treatment: treatment}
		card.Vouchers = append([]Voucher{voucher}, card.Vouchers...)
		card.SolariumPurchases = 0
		return card, &voucher, nil
	}
	return card, nil, nil
}

func (card Card) UseVoucher(voucherID string) (Card, error) {
	for idx, voucher := range card.Vouchers {
		if voucher.ID == voucherID {
			if voucher.IsUsed {
				return Card{}, fmt.Errorf("voucher already used")
			}
			card.Vouchers[idx].IsUsed = true
			return card, nil
		}
	}
	return Card{}, fmt.Errorf("voucher not found")
}
