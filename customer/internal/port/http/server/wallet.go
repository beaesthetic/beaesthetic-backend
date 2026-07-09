package server

import (
	"context"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/application"
	customerdomain "github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/customer"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/wallet"
	walletapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/wallet"
)

func (s *Server) GetWallets(ctx context.Context, request walletapi.GetWalletsRequestObject) (walletapi.GetWalletsResponseObject, error) {
	wallets, err := s.wallet.GetAll(ctx, stringValue(request.Params.Filter))
	if err != nil {
		return nil, err
	}
	out := make([]walletapi.Wallet, 0, len(wallets))
	for _, item := range wallets {
		out = append(out, walletResponse(item))
	}
	return walletapi.GetWallets200JSONResponse(out), nil
}

func (s *Server) AddGiftCard(ctx context.Context, request walletapi.AddGiftCardRequestObject) (walletapi.AddGiftCardResponseObject, error) {
	if request.Body == nil {
		return nil, errMissingBody
	}
	wallet, err := s.wallet.AddGiftCard(ctx, request.Body.CustomerId.String(), float64(request.Body.Amount))
	if err != nil {
		return nil, err
	}
	id := uuidPtr(wallet.ID)
	return walletapi.AddGiftCard200JSONResponse{Id: (*walletapi.WalletId)(id)}, nil
}

func (s *Server) GetWalletById(ctx context.Context, request walletapi.GetWalletByIdRequestObject) (walletapi.GetWalletByIdResponseObject, error) {
	wallet, err := s.wallet.GetByID(ctx, request.WalletId.String())
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return walletapi.GetWalletById404Response{}, nil
	}
	return walletapi.GetWalletById200JSONResponse(walletResponse(*wallet)), nil
}

func (s *Server) ChargeWallet(ctx context.Context, request walletapi.ChargeWalletRequestObject) (walletapi.ChargeWalletResponseObject, error) {
	if request.Body == nil {
		return nil, errMissingBody
	}
	wallet, err := s.wallet.Charge(ctx, request.WalletId.String(), float64(request.Body.Amount))
	if err != nil {
		return nil, err
	}
	id := uuidPtr(wallet.ID)
	return walletapi.ChargeWallet200JSONResponse{Id: (*walletapi.WalletId)(id)}, nil
}

func walletResponse(model application.WalletReadModel) walletapi.Wallet {
	id := uuidPtr(model.Wallet.ID)
	available := float32(model.Wallet.AvailableAmount)
	spent := float32(model.Wallet.Spent)
	history := walletOperations(model.Wallet.Operations)
	customer := walletCustomer(model.Customer)
	return walletapi.Wallet{
		AvailableAmount: &available,
		CreatedAt:       &model.Wallet.CreatedAt,
		Customer:        &customer,
		History:         &history,
		Id:              (*walletapi.WalletId)(id),
		Spent:           &spent,
		UpdatedAt:       &model.Wallet.UpdatedAt,
	}
}

func walletCustomer(customer customerdomain.Customer) walletapi.Customer {
	return walletapi.Customer{
		Email:   emailStringPtr(customer.Email),
		Id:      customer.ID,
		Name:    customer.Name,
		Phone:   phoneStringPtr(customer.Phone),
		Surname: customer.Surname,
	}
}

func walletOperations(ops []wallet.Operation) []walletapi.WalletOperation {
	out := make([]walletapi.WalletOperation, 0, len(ops))
	for _, op := range ops {
		amount := float32(op.Amount)
		operation := walletapi.WalletOperation{}
		switch op.Type {
		case "giftCardMoneyCredited":
			giftCardID := uuidPtr(op.GiftCardID)
			_ = operation.FromGiftCardMoneyCreditedEvent(walletapi.GiftCardMoneyCreditedEvent{Amount: &amount, At: &op.At, ExpireAt: timePtrIfNotZero(op.ExpireAt), GiftCardId: (*walletapi.GiftCardId)(giftCardID)})
		case "giftCardMoneyExpired":
			giftCardID := uuidPtr(op.GiftCardID)
			_ = operation.FromGiftCardMoneyExpiredEvent(walletapi.GiftCardMoneyExpiredEvent{Amount: &amount, At: &op.At, GiftCardId: (*walletapi.GiftCardId)(giftCardID)})
		case "moneyCharged":
			_ = operation.FromMoneyChargedEvent(walletapi.MoneyChargedEvent{Amount: &amount, At: &op.At})
		default:
			_ = operation.FromMoneyCreditedEvent(walletapi.MoneyCreditedEvent{Amount: &amount, At: &op.At})
		}
		out = append(out, operation)
	}
	return out
}
