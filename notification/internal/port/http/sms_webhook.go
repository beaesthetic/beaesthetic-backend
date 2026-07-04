package http

import (
	"context"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/provider"
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
		if request.Body.Metadata == nil {
			return nil, badRequest("missing mandatory metadata")
		}
		notificationID := (*request.Body.Metadata)[provider.NotificationIDMetadata]
		if notificationID == "" {
			return nil, badRequest("missing mandatory metadata")
		}
		if err := server.service.ConfirmNotificationSent(ctx, notificationID); err != nil {
			server.log.Error("failed to confirm notification sent", zap.Error(err))
			return nil, err
		}
		server.log.Info(
			"sms delivery confirmed",
			zap.String("notification_id", notificationID),
			zap.String("event_type", string(*request.Body.EventType)),
		)
	case smswebhook.MessageDeliverFailed:
		server.log.Info("received failed SMS delivery event", zap.String("event_type", string(*request.Body.EventType)))
	default:
		server.log.Warn("unexpected SMS delivery event", zap.String("event_type", string(*request.Body.EventType)))
	}
	return smswebhook.SmsGatewayNotify200Response{}, nil
}
