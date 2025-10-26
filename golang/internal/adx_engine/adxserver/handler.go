package adxserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
	"github.com/echoface/admux/internal/adx_engine/adxcore"
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
	// Extract request context with SSP ID
	ctx := extractRequestContext(c)

	// Get SSP adapter and configuration based on SSP ID
	sspAdapter, sspConfig, err := h.adxServer.GetSSPAdapter(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid SSP configuration",
			"details": err.Error(),
		})
		return
	}

	// Read request body
	bodyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to read request body",
			"details": err.Error(),
		})
		return
	}
	defer c.Request.Body.Close()

	// Create bid request context
	bidCtx := adxcore.NewBidRequestCtx(ctx, nil)
	bidCtx.SetSSPInfo(sspConfig.ID, sspConfig)

	// Convert SSP-specific request to internal format using adapter
	if err := sspAdapter.ToInternalBidRequest(bidCtx, bodyData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to parse bid request",
			"details": err.Error(),
		})
		return
	}

	// Process the bid request through the pipeline
	_, err = h.adxServer.ProcessBid(bidCtx.Context, bidCtx.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process bid request",
			"details": err.Error(),
		})
		return
	}

	// Convert internal response to SSP-specific format
	sspResponse, err := sspAdapter.PackSSPResponse(bidCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to format response",
			"details": err.Error(),
		})
		return
	}

	// Return SSP-specific response
	c.Data(http.StatusOK, "application/json", sspResponse)
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

	// Extract SSP ID from query parameter (preferred) or header
	sspID := c.Query("sspid")
	if sspID == "" {
		sspID = c.GetHeader("X-SSP-ID")
	}
	if sspID == "" {
		// Fallback to ssid for backward compatibility
		sspID = c.Query("ssid")
	}

	if sspID != "" {
		ctx = context.WithValue(ctx, "sspid", sspID)
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
