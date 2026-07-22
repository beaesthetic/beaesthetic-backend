package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/application"
	customerdomain "github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/customer"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/wallet"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/infra/postgres/queries"
)

type WalletRepository struct{ queries *queries.Queries }

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{queries: queries.New(db)}
}
func (r *WalletRepository) Save(ctx context.Context, w wallet.Wallet) (wallet.Wallet, error) {
	operations, err := json.Marshal(w.Operations)
	if err != nil {
		return wallet.Wallet{}, err
	}
	giftCards, err := json.Marshal(w.GiftCards)
	if err != nil {
		return wallet.Wallet{}, err
	}
	return w, r.queries.SaveWallet(ctx, queries.SaveWalletParams{
		ID:              w.ID,
		Owner:           w.Owner,
		AvailableAmount: w.AvailableAmount,
		Spent:           w.Spent,
		Operations:      operations,
		GiftCards:       giftCards,
		CreatedAt:       w.CreatedAt,
		UpdatedAt:       w.UpdatedAt,
	})
}
func (r *WalletRepository) FindDomainByID(ctx context.Context, id string) (*wallet.Wallet, error) {
	row, err := r.queries.FindWalletByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	w := mapWallet(row)
	return &w, nil
}
func (r *WalletRepository) FindByCustomer(ctx context.Context, customerID string) (*wallet.Wallet, error) {
	row, err := r.queries.FindWalletByCustomerID(ctx, customerID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	w := mapWallet(row)
	return &w, nil
}
func (r *WalletRepository) FindByID(ctx context.Context, id string) (*application.WalletReadModel, error) {
	row, err := r.queries.FindWalletReadModelByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	model := mapWalletReadModel(row.ID, row.Owner, row.AvailableAmount, row.Spent, row.Operations, row.GiftCards, row.CreatedAt, row.UpdatedAt, row.CustomerID, row.Name, row.Surname, row.Email, row.Phone, row.Note)
	return &model, nil
}
func (r *WalletRepository) FindAll(ctx context.Context, filter string) ([]application.WalletReadModel, error) {
	var out []application.WalletReadModel
	if strings.TrimSpace(filter) != "" {
		rows, err := r.queries.SearchWalletReadModels(ctx, sql.NullString{String: strings.ToLower(filter), Valid: true})
		if err != nil {
			return nil, err
		}
		out = make([]application.WalletReadModel, 0, len(rows))
		for _, row := range rows {
			out = append(out, mapWalletReadModel(row.ID, row.Owner, row.AvailableAmount, row.Spent, row.Operations, row.GiftCards, row.CreatedAt, row.UpdatedAt, row.CustomerID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
		}
		return out, nil
	}
	rows, err := r.queries.FindWalletReadModels(ctx)
	if err != nil {
		return nil, err
	}
	out = make([]application.WalletReadModel, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapWalletReadModel(row.ID, row.Owner, row.AvailableAmount, row.Spent, row.Operations, row.GiftCards, row.CreatedAt, row.UpdatedAt, row.CustomerID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
	}
	return out, nil
}

func mapWallet(row queries.Wallet) wallet.Wallet {
	w := wallet.Wallet{
		ID:              row.ID,
		Owner:           row.Owner,
		AvailableAmount: row.AvailableAmount,
		Spent:           row.Spent,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	_ = json.Unmarshal(row.Operations, &w.Operations)
	_ = json.Unmarshal(row.GiftCards, &w.GiftCards)
	return w
}

func mapWalletReadModel(id string, owner string, availableAmount float64, spent float64, operations []byte, giftCards []byte, createdAt time.Time, updatedAt time.Time, customerID string, name string, surname string, email sql.NullString, phone sql.NullString, note string) application.WalletReadModel {
	m := application.WalletReadModel{
		Wallet: wallet.Wallet{
			ID:              id,
			Owner:           owner,
			AvailableAmount: availableAmount,
			Spent:           spent,
			CreatedAt:       createdAt,
			UpdatedAt:       updatedAt,
		},
		Customer: customerdomain.Customer{
			ID:      customerID,
			Name:    name,
			Surname: surname,
			Note:    note,
		},
	}
	_ = json.Unmarshal(operations, &m.Wallet.Operations)
	_ = json.Unmarshal(giftCards, &m.Wallet.GiftCards)
	if email.Valid {
		m.Customer.Email = &email.String
	}
	if phone.Valid {
		parsed, err := customerdomain.ParsePhone(phone.String)
		if err == nil {
			m.Customer.Phone = parsed
		}
	}
	return m
}
