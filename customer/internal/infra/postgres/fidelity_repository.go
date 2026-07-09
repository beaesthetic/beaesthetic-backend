package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/fidelity"
)

type FidelityRepository struct{ db *sql.DB }

func NewFidelityRepository(db *sql.DB) *FidelityRepository { return &FidelityRepository{db: db} }

func (r *FidelityRepository) Save(ctx context.Context, card fidelity.Card) (fidelity.Card, error) {
	vouchers, err := json.Marshal(card.Vouchers)
	if err != nil {
		return fidelity.Card{}, err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO fidelity_cards (id,customer_id,solarium_purchases,vouchers,updated_at) VALUES ($1,$2,$3,$4,now())
ON CONFLICT (id) DO UPDATE SET customer_id=$2,solarium_purchases=$3,vouchers=$4,updated_at=now()`, card.ID, card.CustomerID, card.SolariumPurchases, vouchers)
	return card, err
}
func (r *FidelityRepository) FindAll(ctx context.Context) ([]fidelity.Card, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,customer_id,solarium_purchases,vouchers FROM fidelity_cards ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCards(rows)
}
func (r *FidelityRepository) FindByID(ctx context.Context, id string) (*fidelity.Card, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,customer_id,solarium_purchases,vouchers FROM fidelity_cards WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cards, err := scanCards(rows)
	if err != nil || len(cards) == 0 {
		return nil, err
	}
	return &cards[0], nil
}
func (r *FidelityRepository) FindByCustomerID(ctx context.Context, customerID string) ([]fidelity.Card, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,customer_id,solarium_purchases,vouchers FROM fidelity_cards WHERE customer_id=$1`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCards(rows)
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
	rows, err := r.db.QueryContext(ctx, `SELECT id,customer_id,solarium_purchases,vouchers FROM fidelity_cards WHERE vouchers @> $1::jsonb LIMIT 1`, string(filter))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cards, err := scanCards(rows)
	if err != nil || len(cards) == 0 {
		return nil, err
	}
	return &cards[0], nil
}

func scanCards(rows *sql.Rows) ([]fidelity.Card, error) {
	out := []fidelity.Card{}
	for rows.Next() {
		var card fidelity.Card
		var vouchers []byte
		if err := rows.Scan(&card.ID, &card.CustomerID, &card.SolariumPurchases, &vouchers); err != nil {
			return nil, err
		}
		if len(vouchers) > 0 {
			_ = json.Unmarshal(vouchers, &card.Vouchers)
		}
		out = append(out, card)
	}
	return out, rows.Err()
}
