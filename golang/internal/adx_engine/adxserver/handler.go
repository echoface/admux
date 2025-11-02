package adxserver

import (
	"context"
	"fmt"
	"io"
	"net/http"

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

func (h *BidHandler) HandleAdMuxBid(c *gin.Context) {
	// Extract request context with SSP ID
	info, err := extractRequestContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrResponse(err, "invalid query params"))
		return
	}

	h.handleBidRequest(c, info)
}

func (h *BidHandler) HandleKuaishouBid(c *gin.Context) {
	// Extract request context with SSP ID
	info, err := extractRequestContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrResponse(err, "invalid query params"))
		return
	}

	h.handleBidRequest(c, info)
}

func (h *BidHandler) handleBidRequest(c *gin.Context, info *reqInfo) {
	// Get SSP adapter and configuration based on SSP ID
	sspAdapter, sspConfig, err := h.adxServer.GetSSPAdapter(info.SSPID)
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrResponse(err, "Invalid SSP configuration"))
		return
	}

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
	bidCtx := adxcore.NewBidRequestCtx(context.Background(), nil)
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
	if err = h.adxServer.ProcessBid(bidCtx); err != nil {
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

type reqInfo struct {
	SSPID    string
	ClientIP string
}

func extractRequestContext(c *gin.Context) (*reqInfo, error) {
	info := &reqInfo{}

	// Extract SSP ID from query parameter (preferred) or header
	info.SSPID = c.Query("sspid")
	if info.SSPID == "" {
		info.SSPID = c.GetHeader("X-SSP-ID")
	}
	if info.SSPID == "" {
		// Fallback to ssid for backward compatibility
		info.SSPID = c.Query("ssid")
	}
	if info.SSPID == "" {
		return nil, fmt.Errorf("mising sspid in query")
	}

	info.ClientIP = c.ClientIP()
	return info, nil
}

func newErrResponse(err error, msg string) gin.H {
	return gin.H{
		"error":   msg,
		"details": err.Error(),
	}
}
