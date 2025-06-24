package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/portal26/event_receiver_service/internal/batch"
	"github.com/portal26/event_receiver_service/internal/config"
	"github.com/portal26/event_receiver_service/internal/model"
)

func TestIngestEvent(t *testing.T) {
	cfg := &config.Config{
		ValidTiers: []string{"pro", "enterprise", "free"},
	}
	bp := &batch.BatchProcessor{events: make(chan model.Event, 1)}
	h := NewHandler(cfg, bp)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/ingest", h.IngestEvent)

	event := model.Event{
		EventTimestamp: time.Now(),
		Body:           "test event",
	}
	eventData, _ := json.Marshal(event)

	req, _ := http.NewRequest("POST", "/ingest", bytes.NewBuffer(eventData))
	req.Header.Set("X-Customer-Tier", "pro")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.Response
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "success", resp.Status)
}

func TestIngestInvalidTier(t *testing.T) {
	cfg := &config.Config{
		ValidTiers: []string{"pro", "enterprise", "free"},
	}
	bp := &batch.BatchProcessor{}
	h := NewHandler(cfg, bp)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/ingest", h.IngestEvent)

	req, _ := http.NewRequest("POST", "/ingest", bytes.NewBuffer([]byte(`{"event_timestamp":"2024-01-01T00:00:00Z","body":"test"}`)))
	req.Header.Set("X-Customer-Tier", "invalid")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
