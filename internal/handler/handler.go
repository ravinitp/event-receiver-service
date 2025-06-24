package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/portal26/event_receiver_service/internal/batch"
	"github.com/portal26/event_receiver_service/internal/config"
	"github.com/portal26/event_receiver_service/internal/metrics"
	"github.com/portal26/event_receiver_service/internal/model"
)

type Handler struct {
	cfg            *config.Config
	batchProcessor *batch.BatchProcessor
}

func NewHandler(cfg *config.Config, bp *batch.BatchProcessor) *Handler {
	return &Handler{cfg: cfg, batchProcessor: bp}
}

func (h *Handler) IngestEvent(c *gin.Context) {
	metrics.RequestsTotal.Inc()

	tier := c.GetHeader("X-Customer-Tier")
	if !contains(h.cfg.ValidTiers, tier) {
		metrics.FilteredRequests.Inc()
		log.Printf("Invalid customer tier: %s", tier)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer tier"})
		return
	}

	var event model.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		metrics.ErrorsTotal.Inc()
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	h.batchProcessor.AddEvent(event)
	metrics.EventsProcessed.Inc()
	log.Printf("Event processed: %s", event.EventTimestamp)

	c.JSON(http.StatusOK, model.Response{Status: "success"})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
