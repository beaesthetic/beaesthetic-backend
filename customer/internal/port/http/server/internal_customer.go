package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	customerapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/customer"
	"go.uber.org/zap"
)

// internalCustomerResponse deliberately excludes administrative-only fields,
// such as the customer note. It is the contract used by private callers.
type internalCustomerResponse struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Surname string  `json:"surname"`
	Email   *string `json:"email,omitempty"`
	Phone   *string `json:"phone,omitempty"`
}

func internalCustomerByID(handler customerapi.StrictServerInterface, log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response, err := handler.GetCustomerById(ctx, customerapi.GetCustomerByIdRequestObject{CustomerId: ctx.Param("customerId")})
		if err != nil {
			status := http.StatusInternalServerError
			if isNotFound(err) {
				status = http.StatusNotFound
			}
			log.Error("internal customer lookup failed", zap.Error(err), zap.String("customer_id", ctx.Param("customerId")))
			ctx.JSON(status, gin.H{"message": err.Error()})
			return
		}

		customer, ok := response.(customerapi.GetCustomerById200JSONResponse)
		if !ok {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "customer not found"})
			return
		}
		ctx.JSON(http.StatusOK, internalCustomerResponse{
			ID:      customer.Id,
			Name:    customer.Name,
			Surname: customer.Surname,
			Email:   (*string)(customer.Email),
			Phone:   customer.Phone,
		})
	}
}
