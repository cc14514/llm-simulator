package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/cc14514/llm-simulator/internal/model"
	"github.com/cc14514/llm-simulator/internal/simulator"
)

// Handler holds HTTP handler methods for the LLM simulator.
type Handler struct {
	sim *simulator.Simulator
}

// New creates a new Handler.
func New(sim *simulator.Simulator) *Handler {
	return &Handler{sim: sim}
}

// Health handles GET /health.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ListModels handles GET /v1/models.
func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.sim.GetModels())
}

// ChatCompletions handles POST /v1/chat/completions.
func (h *Handler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}

	var req model.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "invalid_request_error")
		return
	}

	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "messages is required and must be non-empty", "invalid_request_error")
		return
	}

	// Simulate error injection
	if h.sim.Config.ErrorRate > 0 && rand.Float64() < h.sim.Config.ErrorRate {
		statusCode := h.sim.Config.ErrorStatusCode
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
		writeError(w, statusCode, "simulated error", "server_error")
		return
	}

	// Simulate response delay
	if h.sim.Config.ResponseDelay > 0 {
		time.Sleep(h.sim.Config.ResponseDelay)
	}

	if req.Stream {
		h.handleStreamingResponse(w, req)
		return
	}

	h.handleNonStreamingResponse(w, req)
}

func (h *Handler) handleNonStreamingResponse(w http.ResponseWriter, req model.ChatCompletionRequest) {
	resp := h.sim.GenerateResponse(req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleStreamingResponse(w http.ResponseWriter, req model.ChatCompletionRequest) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported", "server_error")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	chunks := h.sim.GenerateStreamChunks(req)
	for _, chunk := range chunks {
		data, err := json.Marshal(chunk)
		if err != nil {
			log.Printf("error marshaling chunk: %v", err)
			return
		}
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		if h.sim.Config.StreamChunkDelay > 0 {
			time.Sleep(h.sim.Config.StreamChunkDelay)
		}
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func writeError(w http.ResponseWriter, statusCode int, message, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error: model.ErrorDetail{
			Message: message,
			Type:    errType,
		},
	})
}
