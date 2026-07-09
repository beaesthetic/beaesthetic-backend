package server

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"
	customerdomain "github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/customer"
	customerapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/customer"
)

func (s *Server) GetAllCustomers(ctx context.Context, request customerapi.GetAllCustomersRequestObject) (customerapi.GetAllCustomersResponseObject, error) {
	items, err := s.customers.Search(ctx, stringValue(request.Params.Filter), intValue(request.Params.Limit, 50))
	if err != nil {
		return nil, err
	}
	return customerapi.GetAllCustomers200JSONResponse(customerResponses(items)), nil
}

func (s *Server) CreateCustomer(ctx context.Context, request customerapi.CreateCustomerRequestObject) (customerapi.CreateCustomerResponseObject, error) {
	if request.Body == nil {
		return nil, errMissingBody
	}
	customer, err := s.customers.Create(ctx, request.Body.Name, stringValue(request.Body.Surname), emailPtrToStringPtr(request.Body.Email), request.Body.Phone, request.Body.Note)
	if err != nil {
		return nil, err
	}
	return customerapi.CreateCustomer201JSONResponse{Id: &customer.ID}, nil
}

func (s *Server) GetCustomerByPage(ctx context.Context, request customerapi.GetCustomerByPageRequestObject) (customerapi.GetCustomerByPageResponseObject, error) {
	items, next, hasNext, hasPrevious, err := s.customers.Page(ctx, stringValue(request.Params.PageToken), intValue(request.Params.Limit, 50), stringValue(request.Params.SortBy), string(request.Direction))
	if err != nil {
		return nil, err
	}
	responses := customerResponses(items)
	itemCount := float32(len(responses))
	return customerapi.GetCustomerByPage200JSONResponse{
		HasNextPage:     &hasNext,
		HasPreviousPage: &hasPrevious,
		ItemCount:       &itemCount,
		Items:           &responses,
		NextCursor:      stringPtrIfNotEmpty(next),
	}, nil
}

func (s *Server) DeleteCustomer(ctx context.Context, request customerapi.DeleteCustomerRequestObject) (customerapi.DeleteCustomerResponseObject, error) {
	deleted, err := s.customers.Delete(ctx, request.CustomerId)
	if err != nil {
		return nil, err
	}
	if !deleted {
		return nil, errNotFound("customer")
	}
	return customerapi.DeleteCustomer204Response{}, nil
}

func (s *Server) GetCustomerById(ctx context.Context, request customerapi.GetCustomerByIdRequestObject) (customerapi.GetCustomerByIdResponseObject, error) {
	customer, err := s.customers.Get(ctx, request.CustomerId)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, errNotFound("customer")
	}
	return customerapi.GetCustomerById200JSONResponse(customerResponse(*customer)), nil
}

func (s *Server) UpdateCustomerById(ctx context.Context, request customerapi.UpdateCustomerByIdRequestObject) (customerapi.UpdateCustomerByIdResponseObject, error) {
	if request.Body == nil {
		return nil, errMissingBody
	}
	customer, err := s.customers.Update(ctx, request.CustomerId, request.Body.Name, request.Body.Surname, emailPtrToStringPtr(request.Body.Email), request.Body.Phone, request.Body.Note)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, errNotFound("customer")
	}
	return customerapi.UpdateCustomerById200JSONResponse(customerResponse(*customer)), nil
}

func (s *Server) SearchCustomer(ctx context.Context, request customerapi.SearchCustomerRequestObject) (customerapi.SearchCustomerResponseObject, error) {
	items, err := s.customers.Search(ctx, stringValue(request.Params.Filter), intValue(request.Params.Limit, 50))
	if err != nil {
		return nil, err
	}
	return customerapi.SearchCustomer200JSONResponse(customerResponses(items)), nil
}

func (s *Server) SearchCustomerByPhone(ctx context.Context, request customerapi.SearchCustomerByPhoneRequestObject) (customerapi.SearchCustomerByPhoneResponseObject, error) {
	if request.Body == nil {
		return nil, errMissingBody
	}
	customer, err := s.customers.SearchByPhone(ctx, request.Body.Phone)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return customerapi.SearchCustomerByPhone404Response{}, nil
	}
	return customerapi.SearchCustomerByPhone200JSONResponse(customerResponse(*customer)), nil
}

func customerResponse(customer customerdomain.Customer) customerapi.CustomerResponse {
	return customerapi.CustomerResponse{
		Email:   emailStringPtr(customer.Email),
		Id:      customer.ID,
		Name:    customer.Name,
		Note:    stringPtrIfNotEmpty(customer.Note),
		Phone:   phoneStringPtr(customer.Phone),
		Surname: customer.Surname,
	}
}

func customerResponses(items []customerdomain.Customer) []customerapi.CustomerResponse {
	out := make([]customerapi.CustomerResponse, 0, len(items))
	for _, item := range items {
		out = append(out, customerResponse(item))
	}
	return out
}

func emailPtrToStringPtr(value *openapi_types.Email) *string {
	if value == nil {
		return nil
	}
	out := string(*value)
	return &out
}

func emailStringPtr(value *string) *openapi_types.Email {
	if value == nil {
		return nil
	}
	out := openapi_types.Email(*value)
	return &out
}

func phoneStringPtr(value *customerdomain.Phone) *string {
	if value == nil {
		return nil
	}
	out := value.FullNumber()
	return &out
}
