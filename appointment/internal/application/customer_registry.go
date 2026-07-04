package application

import "context"

type Customer struct {
	ID          string
	DisplayName string
	PhoneNumber *string
}

type CustomerRegistry interface {
	FindByCustomerID(ctx context.Context, customerID string) (*Customer, error)
}
