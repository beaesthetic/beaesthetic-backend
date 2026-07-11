package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/wallet"
)

func TestWalletOperationsIncludeDiscriminatorType(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 10, 18, 14, 7, 0, time.UTC)
	operations := walletOperations([]wallet.Operation{
		{Type: "giftCardMoneyCredited", Amount: 12, At: now, GiftCardID: "696be406-2a04-4de6-b052-8c67d7c15d32", ExpireAt: now.Add(365 * 24 * time.Hour)},
		{Type: "giftCardMoneyExpired", Amount: 10, At: now, GiftCardID: "696be406-2a04-4de6-b052-8c67d7c15d32"},
		{Type: "moneyCharged", Amount: 5, At: now},
		{Type: "moneyCredited", Amount: 3, At: now},
	})

	body, err := json.Marshal(operations)
	if err != nil {
		t.Fatalf("marshal wallet operations: %v", err)
	}

	var payload []map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal wallet operations: %v", err)
	}

	want := []string{"GiftCardMoneyCredited", "GiftCardMoneyExpired", "MoneyCharged", "MoneyCredited"}
	if len(payload) != len(want) {
		t.Fatalf("operations = %d, want %d", len(payload), len(want))
	}
	for i, expected := range want {
		if payload[i]["type"] != expected {
			t.Fatalf("operation %d type = %#v, want %q; json=%s", i, payload[i]["type"], expected, body)
		}
	}
}
