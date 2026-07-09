package application

import (
	"context"
	"time"

	customerdomain "github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/customer"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/wallet"
)

type WalletRepository interface {
	Save(ctx context.Context, wallet wallet.Wallet) (wallet.Wallet, error)
	FindAll(ctx context.Context, filter string) ([]WalletReadModel, error)
	FindByID(ctx context.Context, id string) (*WalletReadModel, error)
	FindDomainByID(ctx context.Context, id string) (*wallet.Wallet, error)
	FindByCustomer(ctx context.Context, customerID string) (*wallet.Wallet, error)
}

type WalletReadModel struct {
	Wallet   wallet.Wallet
	Customer customerdomain.Customer
}

type WalletService struct{ repo WalletRepository }

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) AddGiftCard(ctx context.Context, customerID string, amount float64) (wallet.Wallet, error) {
	now := time.Now().UTC()
	w, err := s.repo.FindByCustomer(ctx, customerID)
	if err != nil {
		return wallet.Wallet{}, err
	}
	if w == nil {
		created := wallet.New(customerID, now)
		w = &created
	}
	updated, err := w.CreditGiftCard(amount, now)
	if err != nil {
		return wallet.Wallet{}, err
	}
	return s.repo.Save(ctx, updated)
}

func (s *WalletService) Charge(ctx context.Context, walletID string, amount float64) (wallet.Wallet, error) {
	w, err := s.repo.FindDomainByID(ctx, walletID)
	if err != nil || w == nil {
		return wallet.Wallet{}, err
	}
	updated, err := w.Charge(amount, time.Now().UTC())
	if err != nil {
		return wallet.Wallet{}, err
	}
	return s.repo.Save(ctx, updated)
}

func (s *WalletService) GetByID(ctx context.Context, walletID string) (*WalletReadModel, error) {
	return s.repo.FindByID(ctx, walletID)
}

func (s *WalletService) GetAll(ctx context.Context, filter string) ([]WalletReadModel, error) {
	return s.repo.FindAll(ctx, filter)
}
