package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"go.uber.org/zap"
)

type Server struct {
	customerNotificationService *application.CustomerNotificationService
	log                         *zap.Logger
}

var _ smswebhook.StrictServerInterface = (*Server)(nil)

func NewSmsWebhookHandler(customerNotificationService *application.CustomerNotificationService, log *zap.Logger) *Server {
	if log == nil {
		log = zap.NewNop()
	}
	return &Server{customerNotificationService: customerNotificationService, log: log}
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
