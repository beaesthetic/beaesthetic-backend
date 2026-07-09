package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/application"
	customerdomain "github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/customer"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/wallet"
)

type WalletRepository struct{ db *sql.DB }

func NewWalletRepository(db *sql.DB) *WalletRepository { return &WalletRepository{db: db} }
func (r *WalletRepository) Save(ctx context.Context, w wallet.Wallet) (wallet.Wallet, error) {
	operations, err := json.Marshal(w.Operations)
	if err != nil {
		return wallet.Wallet{}, err
	}
	giftCards, err := json.Marshal(w.GiftCards)
	if err != nil {
		return wallet.Wallet{}, err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO wallets (id,owner,available_amount,spent,operations,gift_cards,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT (id) DO UPDATE SET owner=$2,available_amount=$3,spent=$4,operations=$5,gift_cards=$6,updated_at=$8`, w.ID, w.Owner, w.AvailableAmount, w.Spent, operations, giftCards, w.CreatedAt, w.UpdatedAt)
	return w, err
}
func (r *WalletRepository) FindDomainByID(ctx context.Context, id string) (*wallet.Wallet, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,owner,available_amount,spent,operations,gift_cards,created_at,updated_at FROM wallets WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	wallets, err := scanWallets(rows)
	if err != nil || len(wallets) == 0 {
		return nil, err
	}
	return &wallets[0], nil
}
func (r *WalletRepository) FindByCustomer(ctx context.Context, customerID string) (*wallet.Wallet, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,owner,available_amount,spent,operations,gift_cards,created_at,updated_at FROM wallets WHERE owner=$1`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	wallets, err := scanWallets(rows)
	if err != nil || len(wallets) == 0 {
		return nil, err
	}
	return &wallets[0], nil
}
func (r *WalletRepository) FindByID(ctx context.Context, id string) (*application.WalletReadModel, error) {
	rows, err := r.db.QueryContext(ctx, walletReadQuery()+` WHERE w.id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	models, err := scanWalletReadModels(rows)
	if err != nil || len(models) == 0 {
		return nil, err
	}
	return &models[0], nil
}
func (r *WalletRepository) FindAll(ctx context.Context, filter string) ([]application.WalletReadModel, error) {
	query := walletReadQuery()
	args := []any{}
	if strings.TrimSpace(filter) != "" {
		query += ` WHERE c.search_text ILIKE '%' || $1 || '%' OR c.search_text % $1`
		args = append(args, strings.ToLower(filter))
	}
	query += ` ORDER BY w.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWalletReadModels(rows)
}
func walletReadQuery() string {
	return `SELECT w.id,w.owner,w.available_amount,w.spent,w.operations,w.gift_cards,w.created_at,w.updated_at,c.id,c.name,c.surname,c.email,c.phone,c.note FROM wallets w JOIN customers c ON c.id=w.owner`
}
func scanWallets(rows *sql.Rows) ([]wallet.Wallet, error) {
	out := []wallet.Wallet{}
	for rows.Next() {
		var w wallet.Wallet
		var operations, giftCards []byte
		if err := rows.Scan(&w.ID, &w.Owner, &w.AvailableAmount, &w.Spent, &operations, &giftCards, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(operations, &w.Operations)
		_ = json.Unmarshal(giftCards, &w.GiftCards)
		out = append(out, w)
	}
	return out, rows.Err()
}
func scanWalletReadModels(rows *sql.Rows) ([]application.WalletReadModel, error) {
	out := []application.WalletReadModel{}
	for rows.Next() {
		var m application.WalletReadModel
		var operations, giftCards []byte
		var email, phone *string
		if err := rows.Scan(&m.Wallet.ID, &m.Wallet.Owner, &m.Wallet.AvailableAmount, &m.Wallet.Spent, &operations, &giftCards, &m.Wallet.CreatedAt, &m.Wallet.UpdatedAt, &m.Customer.ID, &m.Customer.Name, &m.Customer.Surname, &email, &phone, &m.Customer.Note); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(operations, &m.Wallet.Operations)
		_ = json.Unmarshal(giftCards, &m.Wallet.GiftCards)
		m.Customer.Email = email
		if phone != nil {
			parsed, err := customerdomain.ParsePhone(*phone)
			if err == nil {
				m.Customer.Phone = parsed
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
