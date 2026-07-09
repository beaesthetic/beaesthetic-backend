package wallet

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const giftCardDuration = 365 * 24 * time.Hour

type Wallet struct {
	ID              string
	Owner           string
	AvailableAmount float64
	Spent           float64
	Operations      []Operation
	GiftCards       []GiftCard
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type GiftCard struct {
	ID              string    `json:"id"`
	Owner           string    `json:"owner"`
	AvailableAmount float64   `json:"availableAmount"`
	CreatedAt       time.Time `json:"createdAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
	AmountSpent     float64   `json:"amountSpent"`
}

type Operation struct {
	Type       string    `json:"type"`
	Amount     float64   `json:"amount,omitempty"`
	At         time.Time `json:"at"`
	GiftCardID string    `json:"giftCardId,omitempty"`
	ExpireAt   time.Time `json:"expireAt,omitempty"`
}

func New(owner string, now time.Time) Wallet {
	return Wallet{ID: uuid.NewString(), Owner: owner, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
}

func (w Wallet) CreditGiftCard(amount float64, now time.Time) (Wallet, error) {
	if amount <= 0 {
		return Wallet{}, fmt.Errorf("amount must be positive")
	}
	card := GiftCard{ID: uuid.NewString(), Owner: w.Owner, AvailableAmount: amount, CreatedAt: now.UTC(), ExpiresAt: now.Add(giftCardDuration).UTC()}
	w.AvailableAmount += amount
	w.GiftCards = append([]GiftCard{card}, w.GiftCards...)
	w.Operations = append([]Operation{{Type: "giftCardMoneyCredited", Amount: amount, At: now.UTC(), GiftCardID: card.ID, ExpireAt: card.ExpiresAt}}, w.Operations...)
	w.UpdatedAt = now.UTC()
	return w, nil
}

func (w Wallet) Charge(amount float64, now time.Time) (Wallet, error) {
	if amount <= 0 {
		return Wallet{}, fmt.Errorf("amount must be positive")
	}
	w = w.removeExpiredGiftCards(now)
	if amount > w.AvailableAmount {
		return Wallet{}, fmt.Errorf("cannot redeem %.2f cause exceed maximum amount of %.2f", amount, w.AvailableAmount)
	}
	remaining := amount
	for idx := range w.GiftCards {
		if remaining <= 0 {
			break
		}
		card := &w.GiftCards[idx]
		if card.ExpiresAt.Before(now) {
			continue
		}
		charge := min(card.AvailableAmount, remaining)
		card.AvailableAmount -= charge
		card.AmountSpent += charge
		remaining -= charge
	}
	w.AvailableAmount -= amount
	w.Spent += amount
	w.Operations = append([]Operation{{Type: "moneyCharged", Amount: amount, At: now.UTC()}}, w.Operations...)
	w.UpdatedAt = now.UTC()
	return w, nil
}

func (w Wallet) removeExpiredGiftCards(now time.Time) Wallet {
	active := make([]GiftCard, 0, len(w.GiftCards))
	for _, card := range w.GiftCards {
		if card.ExpiresAt.Before(now) && card.AvailableAmount > 0 {
			w.AvailableAmount -= card.AvailableAmount
			w.Operations = append([]Operation{{Type: "giftCardMoneyExpired", Amount: card.AvailableAmount, At: now.UTC(), GiftCardID: card.ID}}, w.Operations...)
			continue
		}
		active = append(active, card)
	}
	w.GiftCards = active
	return w
}
