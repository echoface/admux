package adxserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
	"github.com/gin-gonic/gin"
)

type BidHandler struct {
	adxServer *AdxServer
	appCtx    *AdxServerContext
}

func NewBidHandler(adxServer *AdxServer, appCtx *AdxServerContext) *BidHandler {
	return &BidHandler{
		adxServer: adxServer,
		appCtx:    appCtx,
	}
}

func (h *BidHandler) HandleBidRequest(c *gin.Context) {
	// Extract request context
	ctx := extractRequestContext(c)

	// Parse request body as OpenRTB BidRequest
	bidReq, err := parseBidRequest(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid bid request format",
			"details": err.Error(),
		})
		return
	}

	// Process the bid request through the pipeline
	response, err := h.adxServer.ProcessBid(ctx, bidReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process bid request",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// parseBidRequest parses the request body into an OpenRTB BidRequest
func parseBidRequest(body io.ReadCloser) (*admux_rtb.BidRequest, error) {
	defer body.Close()

	// Read the request body
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	// Parse as JSON into OpenRTB BidRequest
	bidReq := &admux_rtb.BidRequest{}
	if err := json.Unmarshal(data, bidReq); err != nil {
		return nil, err
	}

	return bidReq, nil
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
