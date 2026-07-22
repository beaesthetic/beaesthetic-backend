package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/fidelity"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/infra/postgres/queries"
)

type FidelityRepository struct{ queries *queries.Queries }

func NewFidelityRepository(db *sql.DB) *FidelityRepository {
	return &FidelityRepository{queries: queries.New(db)}
}

func (r *FidelityRepository) Save(ctx context.Context, card fidelity.Card) (fidelity.Card, error) {
	vouchers, err := json.Marshal(card.Vouchers)
	if err != nil {
		return fidelity.Card{}, err
	}
	return card, r.queries.SaveFidelityCard(ctx, queries.SaveFidelityCardParams{
		ID:                card.ID,
		CustomerID:        card.CustomerID,
		SolariumPurchases: int32(card.SolariumPurchases),
		Vouchers:          vouchers,
	})
}
func (r *FidelityRepository) FindAll(ctx context.Context) ([]fidelity.Card, error) {
	rows, err := r.queries.FindFidelityCards(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]fidelity.Card, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapFidelityCard(row.ID, row.CustomerID, row.SolariumPurchases, row.Vouchers))
	}
	return out, nil
}
func (r *FidelityRepository) FindByID(ctx context.Context, id string) (*fidelity.Card, error) {
	row, err := r.queries.FindFidelityCardByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	card := mapFidelityCard(row.ID, row.CustomerID, row.SolariumPurchases, row.Vouchers)
	return &card, nil
}
func (r *FidelityRepository) FindByCustomerID(ctx context.Context, customerID string) ([]fidelity.Card, error) {
	rows, err := r.queries.FindFidelityCardsByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	out := make([]fidelity.Card, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapFidelityCard(row.ID, row.CustomerID, row.SolariumPurchases, row.Vouchers))
	}
	return out, nil
}
func (r *FidelityRepository) FindOneByCustomerID(ctx context.Context, customerID string) (*fidelity.Card, error) {
	cards, err := r.FindByCustomerID(ctx, customerID)
	if err != nil || len(cards) == 0 {
		return nil, err
	}
	return &cards[0], nil
}
func (r *FidelityRepository) FindByVoucherID(ctx context.Context, voucherID string) (*fidelity.Card, error) {
	filter, err := json.Marshal([]map[string]string{{"id": voucherID}})
	if err != nil {
		return nil, err
	}
	row, err := r.queries.FindFidelityCardByVoucherID(ctx, filter)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	card := mapFidelityCard(row.ID, row.CustomerID, row.SolariumPurchases, row.Vouchers)
	return &card, nil
}

func mapFidelityCard(id string, customerID string, solariumPurchases int32, vouchers []byte) fidelity.Card {
	card := fidelity.Card{ID: id, CustomerID: customerID, SolariumPurchases: int(solariumPurchases)}
	if len(vouchers) > 0 {
		_ = json.Unmarshal(vouchers, &card.Vouchers)
	}
	return card
}
