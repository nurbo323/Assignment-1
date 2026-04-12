package http

import (
    "errors"
    "net/http"
    "order-service/internal/domain"
    "order-service/internal/usecase"

    "github.com/gin-gonic/gin"
)

type OrderHandler struct {
    uc usecase.OrderService
}

func NewOrderHandler(uc usecase.OrderService) *OrderHandler {
    return &OrderHandler{uc: uc}
}

type createOrderRequest struct {
    CustomerID string `json:"customer_id" binding:"required"`
    ItemName   string `json:"item_name" binding:"required"`
    Amount     int64  `json:"amount" binding:"required"`
}

func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
    r.POST("/orders", h.CreateOrder)
    r.GET("/orders/:id", h.GetOrder)
    r.PATCH("/orders/:id/cancel", h.CancelOrder)
    r.GET("/order/stats", h.GetStats)
    r.GET("/orders/stats", h.GetStats)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req createOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }

    order, err := h.uc.CreateOrder(c.Request.Context(), usecase.CreateOrderInput{
        CustomerID: req.CustomerID,
        ItemName:   req.ItemName,
        Amount:     req.Amount,
    })
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrConflictIdempotency):
            c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusServiceUnavailable, gin.H{
                "error": "payment service unavailable",
                "order": order,
            })
        }
        return
    }

    c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
    order, err := h.uc.GetOrder(c.Request.Context(), c.Param("id"))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
    order, err := h.uc.CancelOrder(c.Request.Context(), c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetStats(c *gin.Context) {
    stats, err := h.uc.GetStats(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, stats)
}
