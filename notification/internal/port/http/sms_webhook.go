package http

import (
	"context"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"go.uber.org/zap"
)

func (server *Server) SmsGatewayNotify(ctx context.Context, request smswebhook.SmsGatewayNotifyRequestObject) (smswebhook.SmsGatewayNotifyResponseObject, error) {
	if request.Body == nil {
		return nil, badRequest("missing request body")
	}
	if request.Body.EventType == nil {
		return nil, badRequest("missing eventType")
	}
	switch *request.Body.EventType {
	case smswebhook.MessageDeliverSucceeded:
		if request.Body.Data == nil {
			return nil, badRequest("missing data")
		}
		smsGatewayMessageID := request.Body.Data.Id.String()
		matched, err := server.customerNotificationService.ConfirmSMSGatewayMessageSent(ctx, smsGatewayMessageID)
		if err != nil {
			server.log.Error("failed to confirm customer notification sent", zap.String("sms_gateway_message_id", smsGatewayMessageID), zap.Error(err))
			return nil, err
		}
		if !matched {
			server.log.Warn("customer notification not found while confirming sms delivery", zap.String("sms_gateway_message_id", smsGatewayMessageID))
			return smswebhook.SmsGatewayNotify200Response{}, nil
		}
		server.log.Info("customer notification sms delivery confirmed", zap.String("sms_gateway_message_id", smsGatewayMessageID))
	case smswebhook.MessageDeliverFailed:
		if request.Body.Data == nil {
			return nil, badRequest("missing data")
		}
		smsGatewayMessageID := request.Body.Data.Id.String()
		matched, err := server.customerNotificationService.MarkSMSGatewayMessageFailed(ctx, smsGatewayMessageID)
		if err != nil {
			server.log.Error("failed to mark customer notification failed", zap.String("sms_gateway_message_id", smsGatewayMessageID), zap.Error(err))
			return nil, err
		}
		if !matched {
			server.log.Warn("customer notification not found while marking sms delivery failed", zap.String("sms_gateway_message_id", smsGatewayMessageID))
			return smswebhook.SmsGatewayNotify200Response{}, nil
		}
		server.log.Info("customer notification sms delivery failed", zap.String("sms_gateway_message_id", smsGatewayMessageID))
	default:
		server.log.Warn("unexpected SMS delivery event", zap.String("event_type", string(*request.Body.EventType)))
	}
	return smswebhook.SmsGatewayNotify200Response{}, nil
}
