package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/provider"
	"go.uber.org/zap"
)

type Server struct {
	service *application.NotificationService
	log     *zap.Logger
}

func NewServer(service *application.NotificationService, log *zap.Logger) *Server {
	return &Server{service: service, log: log}
}

func NewRouter(server *Server) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", server.Health)
	api.RegisterHandlers(router, server)
	smswebhook.RegisterHandlers(router, server)
	return router
}

func (server *Server) Health(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

func (server *Server) CreateNotification(ctx *gin.Context) {
	var request api.CreateNotificationJSONRequestBody
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	channel, err := toDomainChannel(request.Channel)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	notification, err := server.service.CreateNotification(ctx.Request.Context(), request.Title, request.Content, channel)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"notificationId": notification.ID})
}

func (server *Server) GetNotification(ctx *gin.Context, notificationID api.NotificationId) {
	notification, err := server.service.GetNotification(ctx.Request.Context(), notificationID.String())
	if err != nil {
		server.log.Error("failed to get notification", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}
	if notification == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "notification not found"})
		return
	}
	ctx.JSON(http.StatusOK, toAPI(*notification))
}

func (server *Server) SmsGatewayNotify(ctx *gin.Context) {
	var request smswebhook.SmsGatewayNotifyJSONRequestBody
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if request.EventType == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "missing eventType"})
		return
	}
	switch *request.EventType {
	case smswebhook.MessageDeliverSucceeded:
		if request.Metadata == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "missing mandatory metadata"})
			return
		}
		notificationID := (*request.Metadata)[provider.NotificationIDMetadata]
		if notificationID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "missing mandatory metadata"})
			return
		}
		if err := server.service.ConfirmNotificationSent(ctx.Request.Context(), notificationID); err != nil {
			server.log.Error("failed to confirm notification sent", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}
	case smswebhook.MessageDeliverFailed:
		server.log.Info("received failed SMS delivery event")
	default:
		server.log.Warn("unexpected SMS delivery event", zap.String("eventType", string(*request.EventType)))
	}
	ctx.Status(http.StatusOK)
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
