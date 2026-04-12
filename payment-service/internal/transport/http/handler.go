package http

import (
	"errors"
	"net/http"
	"payment-service/internal/domain"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct{ uc *usecase.PaymentUseCase }

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler { return &PaymentHandler{uc: uc} }

func (h *PaymentHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/payments", h.CreatePayment)
	router.GET("/payments/:orderID", h.GetByOrderID)
}

type createPaymentRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount" binding:"required"`
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.uc.CreatePayment(c.Request.Context(), usecase.CreatePaymentInput{OrderID: req.OrderID, Amount: req.Amount})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	statusCode := http.StatusCreated
	if payment.Status == domain.PaymentStatusDeclined {
		statusCode = http.StatusOK
	}
	c.JSON(statusCode, payment)
}

func (h *PaymentHandler) GetByOrderID(c *gin.Context) {
	payment, err := h.uc.GetByOrderID(c.Request.Context(), c.Param("orderID"))
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payment)
}
