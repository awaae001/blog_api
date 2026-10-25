package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"blog_api/src/model"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

const (
	defaultStateTTL    = 7 * 24 * time.Hour
	maxStateTTL        = 7 * 24 * time.Hour
	maxStateBodyBytes  = int64(256 * 1024)
	maxStateTotalBytes = int64(8 << 20)
	maxStateEntryCount = 1000
	stateCleanupPeriod = 3 * time.Hour
)

var stateKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// StateResponse describes an in-memory state snapshot.
type StateResponse struct {
	Key       string          `json:"key"`
	Payload   json.RawMessage `json:"payload"`
	ExpiresAt int64           `json:"expires_at"`
}

// StateHandler serves the authenticated in-memory state API.
type StateHandler struct {
	Store *service.StateStore
	now   func() time.Time
}

// NewStateHandler creates a handler with the production state limits.
func NewStateHandler() *StateHandler {
	store := service.NewStateStore(maxStateEntryCount, maxStateBodyBytes, maxStateTotalBytes)
	go store.RunCleanup(context.Background(), stateCleanupPeriod)
	return &StateHandler{
		Store: store,
		now:   time.Now,
	}
}

// PutState handles PUT /api/internal/states/:key.
func (h *StateHandler) PutState(c *gin.Context) {
	key, ok := validStateKey(c)
	if !ok {
		return
	}
	ttl, ok := parseStateTTL(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxStateBodyBytes+1)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, model.ApiResponse{Code: http.StatusRequestEntityTooLarge, Message: "state payload exceeds 256 KiB"})
			return
		}
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: http.StatusBadRequest, Message: "failed to read state payload"})
		return
	}
	if int64(len(payload)) > maxStateBodyBytes {
		c.JSON(http.StatusRequestEntityTooLarge, model.ApiResponse{Code: http.StatusRequestEntityTooLarge, Message: "state payload exceeds 256 KiB"})
		return
	}

	state, created, err := h.Store.Put(key, payload, h.now().Add(ttl))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStatePayload):
			c.JSON(http.StatusBadRequest, model.ApiResponse{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, service.ErrStatePayloadTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, model.ApiResponse{Code: http.StatusRequestEntityTooLarge, Message: err.Error()})
		case errors.Is(err, service.ErrStateCapacityReached):
			c.JSON(http.StatusInsufficientStorage, model.ApiResponse{Code: http.StatusInsufficientStorage, Message: err.Error()})
		case errors.Is(err, service.ErrStateMemoryLimitReached):
			c.JSON(http.StatusInsufficientStorage, model.ApiResponse{Code: http.StatusInsufficientStorage, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: http.StatusInternalServerError, Message: "failed to store state"})
		}
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, model.ApiResponse{Code: status, Message: "success", Data: newStateResponse(key, state)})
}

// GetState handles GET /api/internal/states/:key without mutating its payload.
func (h *StateHandler) GetState(c *gin.Context) {
	key, ok := validStateKey(c)
	if !ok {
		return
	}
	state, found := h.Store.Get(key)
	if !found {
		c.JSON(http.StatusNotFound, model.ApiResponse{Code: http.StatusNotFound, Message: "state not found"})
		return
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: newStateResponse(key, state)})
}

// DeleteState handles DELETE /api/internal/states/:key.
func (h *StateHandler) DeleteState(c *gin.Context) {
	key, ok := validStateKey(c)
	if !ok {
		return
	}
	if !h.Store.Delete(key) {
		c.JSON(http.StatusNotFound, model.ApiResponse{Code: http.StatusNotFound, Message: "state not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func validStateKey(c *gin.Context) (string, bool) {
	key := c.Param("key")
	if !stateKeyPattern.MatchString(key) {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: http.StatusBadRequest, Message: "invalid state key"})
		return "", false
	}
	return key, true
}

func parseStateTTL(c *gin.Context) (time.Duration, bool) {
	rawTTL := c.Query("ttl_seconds")
	if rawTTL == "" {
		return defaultStateTTL, true
	}
	seconds, err := strconv.ParseInt(rawTTL, 10, 64)
	if err != nil || seconds <= 0 || seconds > int64(maxStateTTL/time.Second) {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: http.StatusBadRequest, Message: "ttl_seconds must be between 1 and 604800"})
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}

func newStateResponse(key string, state service.StoredState) StateResponse {
	return StateResponse{
		Key:       key,
		Payload:   state.Payload,
		ExpiresAt: state.ExpiresAt.Unix(),
	}
}
