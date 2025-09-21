package handler

import (
	"context"
	"net/http"

	"github.com/echoface/admux/internal/adxserver"
	"github.com/gin-gonic/gin"
)

type BidHandler struct {
	adxServer *adxserver.AdxServer
}

func NewBidHandler(adxServer *adxserver.AdxServer) *BidHandler {
	return &BidHandler{
		adxServer: adxServer,
	}
}

func (h *BidHandler) HandleBidRequest(c *gin.Context) {
	// Extract request context
	ctx := extractRequestContext(c)

	// Process the bid request through the pipeline
	response, err := h.adxServer.ProcessBid(ctx, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process bid request",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func extractRequestContext(c *gin.Context) context.Context {
	ctx := context.Background()

	// Extract ssid from query parameter
	ssid := c.Query("ssid")
	if ssid != "" {
		ctx = context.WithValue(ctx, "ssid", ssid)
	}

	// Extract client IP
	clientIP := c.ClientIP()
	if clientIP != "" {
		ctx = context.WithValue(ctx, "client_ip", clientIP)
	}

	// Extract other request metadata
	ctx = context.WithValue(ctx, "method", c.Request.Method)
	ctx = context.WithValue(ctx, "path", c.Request.URL.Path)
	ctx = context.WithValue(ctx, "user_agent", c.Request.UserAgent())

	return ctx
}