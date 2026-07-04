package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
	"go.uber.org/zap"
)

type Server struct {
	service *application.NotificationService
	log     *zap.Logger
}

var _ api.StrictServerInterface = (*Server)(nil)
var _ smswebhook.StrictServerInterface = (*Server)(nil)

func NewSmsHandler(service *application.NotificationService, log *zap.Logger) *Server {
	return &Server{service: service, log: log}
}

func (server *Server) CreateNotification(ctx context.Context, request api.CreateNotificationRequestObject) (api.CreateNotificationResponseObject, error) {
	if request.Body == nil {
		return api.CreateNotification400Response{}, nil
	}
	channel, err := toDomainChannel(request.Body.Channel)
	if err != nil {
		return api.CreateNotification400Response{}, nil
	}
	notification, err := server.service.CreateNotification(ctx, request.Body.Title, request.Body.Content, channel)
	if err != nil {
		return api.CreateNotification400Response{}, nil
	}
	server.log.Info(
		"notification created",
		zap.String("notification_id", notification.ID),
		zap.String("channel_type", string(notification.Channel.Type)),
	)
	notificationID := api.NotificationId(uuid.MustParse(notification.ID))
	return api.CreateNotification200JSONResponse{NotificationId: &notificationID}, nil
}

func (server *Server) GetNotification(ctx context.Context, request api.GetNotificationRequestObject) (api.GetNotificationResponseObject, error) {
	notification, err := server.service.GetNotification(ctx, request.NotificationId.String())
	if err != nil {
		server.log.Error("failed to get notification", zap.Error(err))
		return nil, err
	}
	if notification == nil {
		return api.GetNotification404Response{}, nil
	}
	server.log.Info(
		"notification retrieved",
		zap.String("notification_id", notification.ID),
		zap.String("channel_type", string(notification.Channel.Type)),
		zap.Bool("is_sent", notification.IsSent),
		zap.Bool("is_sent_confirmed", notification.IsSentConfirmed),
	)
	return api.GetNotification200JSONResponse(toAPI(*notification)), nil
}

type statusError struct {
	status  int
	message string
}

func (err statusError) Error() string {
	return err.message
}

func badRequest(message string) error {
	return statusError{status: http.StatusBadRequest, message: message}
}

func strictErrorHandler(ctx *gin.Context, err error) {
	var statusErr statusError
	if errors.As(err, &statusErr) {
		ctx.JSON(statusErr.status, gin.H{"message": statusErr.message})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
}

func toDomainChannel(channel api.NotificationChannel) (domain.Channel, error) {
	discriminator, err := channel.Discriminator()
	if err != nil {
		return domain.Channel{}, err
	}
	switch discriminator {
	case "sms":
		sms, err := channel.AsSmsChannel()
		if err != nil {
			return domain.Channel{}, err
		}
		return domain.Channel{Type: domain.ChannelSMS, Phone: sms.Phone}, nil
	case "email":
		email, err := channel.AsEmailChannel()
		if err != nil {
			return domain.Channel{}, err
		}
		return domain.Channel{Type: domain.ChannelEmail, Email: email.Email}, nil
	case "whatsapp":
		whatsapp, err := channel.AsWhatsappChannel()
		if err != nil {
			return domain.Channel{}, err
		}
		return domain.Channel{Type: domain.ChannelWhatsApp, Phone: whatsapp.Phone}, nil
	default:
		return domain.Channel{}, domain.Channel{Type: domain.ChannelType(discriminator)}.Validate()
	}
}

func toAPI(notification domain.Notification) api.Notification {
	channel := api.NotificationChannel{}
	switch notification.Channel.Type {
	case domain.ChannelSMS:
		_ = channel.FromSmsChannel(api.SmsChannel{Phone: notification.Channel.Phone})
	case domain.ChannelEmail:
		_ = channel.FromEmailChannel(api.EmailChannel{Email: notification.Channel.Email})
	case domain.ChannelWhatsApp:
		_ = channel.FromWhatsappChannel(api.WhatsappChannel{Phone: notification.Channel.Phone})
	}
	id := api.NotificationId(uuid.MustParse(notification.ID))
	title := notification.Title
	content := notification.Content
	isSent := notification.IsSent
	isSentConfirmed := notification.IsSentConfirmed
	return api.Notification{
		NotificationId:  &id,
		Title:           &title,
		Content:         &content,
		IsSent:          &isSent,
		IsSentConfirmed: &isSentConfirmed,
		Channel:         &channel,
	}
}
