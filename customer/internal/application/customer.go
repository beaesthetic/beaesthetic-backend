package application

import (
	"context"

	customerdomain "github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/customer"
)

type CustomerRepository interface {
	Save(ctx context.Context, customer customerdomain.Customer) (customerdomain.Customer, error)
	FindByID(ctx context.Context, id string) (*customerdomain.Customer, error)
	FindAll(ctx context.Context, filter string, limit int) ([]customerdomain.Customer, error)
	FindByPhone(ctx context.Context, phone string) (*customerdomain.Customer, error)
	FindPage(ctx context.Context, pageToken string, limit int, sortBy string, direction string) ([]customerdomain.Customer, string, bool, bool, error)
	Delete(ctx context.Context, id string) (bool, error)
}

type CustomerService struct{ repo CustomerRepository }

func NewCustomerService(repo CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) Create(ctx context.Context, name, surname string, email, phone, note *string) (customerdomain.Customer, error) {
	parsedPhone, err := phoneFromPtr(phone)
	if err != nil {
		return customerdomain.Customer{}, err
	}
	customer, err := customerdomain.New(name, surname, email, parsedPhone, valueOrEmpty(note))
	if err != nil {
		return customerdomain.Customer{}, err
	}
	return s.repo.Save(ctx, customer)
}

func (s *CustomerService) Update(ctx context.Context, id string, name, surname, email, phone, note *string) (*customerdomain.Customer, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil || current == nil {
		return current, err
	}
	updated, err := current.Update(name, surname, email, phone, note)
	if err != nil {
		return nil, err
	}
	saved, err := s.repo.Save(ctx, updated)
	return &saved, err
}

func (s *CustomerService) Delete(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}

func (s *CustomerService) Get(ctx context.Context, id string) (*customerdomain.Customer, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CustomerService) Search(ctx context.Context, filter string, limit int) ([]customerdomain.Customer, error) {
	return s.repo.FindAll(ctx, filter, limit)
}

func (s *CustomerService) SearchByPhone(ctx context.Context, phone string) (*customerdomain.Customer, error) {
	return s.repo.FindByPhone(ctx, phone)
}

func (s *CustomerService) Page(ctx context.Context, token string, limit int, sortBy string, direction string) ([]customerdomain.Customer, string, bool, bool, error) {
	return s.repo.FindPage(ctx, token, limit, sortBy, direction)
}

func phoneFromPtr(value *string) (*customerdomain.Phone, error) {
	if value == nil {
		return nil, nil
	}
	return customerdomain.ParsePhone(*value)
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
