package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/fidelity"
	fidelityapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/fidelity"
)

func (s *Server) GetFidelityCards(ctx context.Context, request fidelityapi.GetFidelityCardsRequestObject) (fidelityapi.GetFidelityCardsResponseObject, error) {
	cards, err := s.fidelity.GetAll(ctx)
	if err != nil {
		return fidelityapi.GetFidelityCards400Response{}, nil
	}
	return fidelityapi.GetFidelityCards200JSONResponse(fidelityResponses(cards)), nil
}

func (s *Server) CreateFidelityCard(ctx context.Context, request fidelityapi.CreateFidelityCardRequestObject) (fidelityapi.CreateFidelityCardResponseObject, error) {
	if request.Body == nil {
		return nil, errMissingBody
	}
	card, err := s.fidelity.Create(ctx, request.Body.CustomerId.String())
	if err != nil {
		return nil, err
	}
	return fidelityapi.CreateFidelityCard200JSONResponse(fidelityResponse(card)), nil
}

func (s *Server) GetFidelityCardsByCustomerId(ctx context.Context, request fidelityapi.GetFidelityCardsByCustomerIdRequestObject) (fidelityapi.GetFidelityCardsByCustomerIdResponseObject, error) {
	cards, err := s.fidelity.GetByCustomerID(ctx, request.CustomerId.String())
	if err != nil {
		return nil, err
	}
	return fidelityapi.GetFidelityCardsByCustomerId200JSONResponse(fidelityResponses(cards)), nil
}

func (s *Server) UseVoucher(ctx context.Context, request fidelityapi.UseVoucherRequestObject) (fidelityapi.UseVoucherResponseObject, error) {
	card, err := s.fidelity.UseVoucher(ctx, request.VoucherId.String())
	if err != nil {
		return nil, err
	}
	if card == nil {
		return nil, errNotFound("fidelity card")
	}
	return fidelityapi.UseVoucher200JSONResponse(fidelityResponse(*card)), nil
}

func (s *Server) GetFidelityCardById(ctx context.Context, request fidelityapi.GetFidelityCardByIdRequestObject) (fidelityapi.GetFidelityCardByIdResponseObject, error) {
	card, err := s.fidelity.GetByID(ctx, request.CardId.String())
	if err != nil {
		return nil, err
	}
	if card == nil {
		return fidelityapi.GetFidelityCardById404Response{}, nil
	}
	return fidelityapi.GetFidelityCardById200JSONResponse(fidelityResponse(*card)), nil
}

func (s *Server) NotifyPurchase(ctx context.Context, request fidelityapi.NotifyPurchaseRequestObject) (fidelityapi.NotifyPurchaseResponseObject, error) {
	if request.Body == nil {
		return nil, errMissingBody
	}
	treatment := fidelity.TreatmentSolarium
	if request.Body.Treatment != nil {
		treatment = fidelity.Treatment(*request.Body.Treatment)
	}
	if err := s.fidelity.RegisterPurchase(ctx, request.CardId.String(), treatment); err != nil {
		return nil, err
	}
	return fidelityapi.NotifyPurchase201Response{}, nil
}

func fidelityResponse(card fidelity.Card) fidelityapi.FidelityCardResponse {
	id := uuidPtr(card.ID)
	customer := fidelityapi.Customer{Id: card.CustomerID}
	vouchers := fidelityVouchers(card.Vouchers)
	return fidelityapi.FidelityCardResponse{
		Customer:          &customer,
		Id:                (*fidelityapi.FidelityCardId)(id),
		SolariumPurchases: &card.SolariumPurchases,
		Vouchers:          &vouchers,
	}
}

func fidelityResponses(cards []fidelity.Card) []fidelityapi.FidelityCardResponse {
	out := make([]fidelityapi.FidelityCardResponse, 0, len(cards))
	for _, card := range cards {
		out = append(out, fidelityResponse(card))
	}
	return out
}

func fidelityVouchers(vouchers []fidelity.Voucher) []fidelityapi.Voucher {
	out := make([]fidelityapi.Voucher, 0, len(vouchers))
	for _, voucher := range vouchers {
		id := uuidPtr(voucher.ID)
		treatment := fidelityapi.SupportedVoucherTreatment(voucher.Treatment)
		item := fidelityapi.FreeVoucher{
			Id:        (*fidelityapi.FidelityCardId)(id),
			IssuedAt:  &voucher.IssuedAt,
			IsUsed:    &voucher.IsUsed,
			Treatment: &treatment,
		}
		var apiVoucher fidelityapi.Voucher
		_ = apiVoucher.FromFreeVoucher(item)
		out = append(out, apiVoucher)
	}
	return out
}

func uuidPtr(value string) *uuid.UUID {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return &parsed
}
