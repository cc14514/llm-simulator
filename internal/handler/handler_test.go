package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cc14514/llm-simulator/internal/model"
	"github.com/cc14514/llm-simulator/internal/simulator"
)

func newTestHandler() *Handler {
	cfg := simulator.DefaultConfig()
	cfg.FixedResponse = "Test response"
	cfg.StreamChunkDelay = 0
	sim := simulator.New(cfg)
	return New(sim)
}

func TestHealth(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}
}

func TestListModels(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()

	h.ListModels(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp model.ModelList
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Object != "list" {
		t.Errorf("expected object 'list', got %q", resp.Object)
	}
	if len(resp.Data) == 0 {
		t.Error("expected at least one model")
	}
}

func TestListModels_MethodNotAllowed(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/models", nil)
	w := httptest.NewRecorder()

	h.ListModels(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestChatCompletions_NonStreaming(t *testing.T) {
	h := newTestHandler()
	body := `{"model":"test","messages":[{"role":"user","content":"Hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp model.ChatCompletionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Model != "test" {
		t.Errorf("expected model 'test', got %q", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Test response" {
		t.Errorf("expected 'Test response', got %q", resp.Choices[0].Message.Content)
	}
}

func TestChatCompletions_Streaming(t *testing.T) {
	h := newTestHandler()
	body := `{"model":"test","messages":[{"role":"user","content":"Hello"}],"stream":true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("expected content type 'text/event-stream', got %q", contentType)
	}

	respBody, _ := io.ReadAll(w.Body)
	bodyStr := string(respBody)

	if !strings.Contains(bodyStr, "data: ") {
		t.Error("expected SSE data lines")
	}
	if !strings.Contains(bodyStr, "[DONE]") {
		t.Error("expected [DONE] terminator")
	}
}

func TestChatCompletions_EmptyMessages(t *testing.T) {
	h := newTestHandler()
	body := `{"model":"test","messages":[]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestChatCompletions_InvalidJSON(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestChatCompletions_MethodNotAllowed(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestChatCompletions_ErrorInjection(t *testing.T) {
	cfg := simulator.DefaultConfig()
	cfg.ErrorRate = 1.0 // Always error
	cfg.ErrorStatusCode = 503
	sim := simulator.New(cfg)
	h := New(sim)

	body := `{"model":"test","messages":[{"role":"user","content":"Hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != 503 {
		t.Errorf("expected status 503, got %d", w.Code)
	}

	var errResp model.ErrorResponse
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error.Message != "simulated error" {
		t.Errorf("expected 'simulated error', got %q", errResp.Error.Message)
	}
}
