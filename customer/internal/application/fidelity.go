package application

import (
	"context"
	"errors"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/fidelity"
)

type FidelityRepository interface {
	Save(ctx context.Context, card fidelity.Card) (fidelity.Card, error)
	FindAll(ctx context.Context) ([]fidelity.Card, error)
	FindByID(ctx context.Context, id string) (*fidelity.Card, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]fidelity.Card, error)
	FindOneByCustomerID(ctx context.Context, customerID string) (*fidelity.Card, error)
	FindByVoucherID(ctx context.Context, voucherID string) (*fidelity.Card, error)
}

type FidelityService struct{ repo FidelityRepository }

func NewFidelityService(repo FidelityRepository) *FidelityService {
	return &FidelityService{repo: repo}
}

func (s *FidelityService) Create(ctx context.Context, customerID string) (fidelity.Card, error) {
	existing, err := s.repo.FindOneByCustomerID(ctx, customerID)
	if err != nil {
		return fidelity.Card{}, err
	}
	if existing != nil {
		return fidelity.Card{}, errors.New("fidelity card already exists")
	}
	return s.repo.Save(ctx, fidelity.NewCard(customerID))
}

func (s *FidelityService) GetAll(ctx context.Context) ([]fidelity.Card, error) {
	return s.repo.FindAll(ctx)
}

func (s *FidelityService) GetByID(ctx context.Context, id string) (*fidelity.Card, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *FidelityService) GetByCustomerID(ctx context.Context, customerID string) ([]fidelity.Card, error) {
	return s.repo.FindByCustomerID(ctx, customerID)
}

func (s *FidelityService) RegisterPurchase(ctx context.Context, cardID string, treatment fidelity.Treatment) error {
	card, err := s.repo.FindByID(ctx, cardID)
	if err != nil || card == nil {
		return err
	}
	updated, _, err := card.RegisterPurchase(treatment)
	if err != nil {
		return err
	}
	_, err = s.repo.Save(ctx, updated)
	return err
}

func (s *FidelityService) UseVoucher(ctx context.Context, voucherID string) (*fidelity.Card, error) {
	card, err := s.repo.FindByVoucherID(ctx, voucherID)
	if err != nil || card == nil {
		return card, err
	}
	updated, err := card.UseVoucher(voucherID)
	if err != nil {
		return nil, err
	}
	saved, err := s.repo.Save(ctx, updated)
	return &saved, err
}
