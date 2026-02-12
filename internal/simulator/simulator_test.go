package simulator

import (
	"testing"

	"github.com/cc14514/llm-simulator/internal/model"
)

func TestGenerateResponse_FixedMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FixedResponse = "Hello, world!"
	sim := New(cfg)

	req := model.ChatCompletionRequest{
		Model: "test-model",
		Messages: []model.Message{
			{Role: "user", Content: "Hi"},
		},
	}

	resp := sim.GenerateResponse(req)

	if resp.Model != "test-model" {
		t.Errorf("expected model 'test-model', got %q", resp.Model)
	}
	if resp.Object != "chat.completion" {
		t.Errorf("expected object 'chat.completion', got %q", resp.Object)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message == nil {
		t.Fatal("expected non-nil message")
	}
	if resp.Choices[0].Message.Content != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %q", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].Message.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", resp.Choices[0].Message.Role)
	}
	if resp.Choices[0].FinishReason == nil || *resp.Choices[0].FinishReason != "stop" {
		t.Errorf("expected finish_reason 'stop'")
	}
	if resp.Usage.PromptTokens <= 0 {
		t.Error("expected positive prompt tokens")
	}
	if resp.Usage.CompletionTokens <= 0 {
		t.Error("expected positive completion tokens")
	}
	if resp.Usage.TotalTokens != resp.Usage.PromptTokens+resp.Usage.CompletionTokens {
		t.Error("expected total tokens = prompt + completion")
	}
}

func TestGenerateResponse_EchoMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EchoMode = true
	sim := New(cfg)

	req := model.ChatCompletionRequest{
		Messages: []model.Message{
			{Role: "user", Content: "What is Go?"},
		},
	}

	resp := sim.GenerateResponse(req)

	expected := "Echo: What is Go?"
	if resp.Choices[0].Message.Content != expected {
		t.Errorf("expected %q, got %q", expected, resp.Choices[0].Message.Content)
	}
}

func TestGenerateResponse_DefaultModel(t *testing.T) {
	cfg := DefaultConfig()
	sim := New(cfg)

	req := model.ChatCompletionRequest{
		Messages: []model.Message{
			{Role: "user", Content: "Hi"},
		},
	}

	resp := sim.GenerateResponse(req)

	if resp.Model != cfg.DefaultModel {
		t.Errorf("expected model %q, got %q", cfg.DefaultModel, resp.Model)
	}
}

func TestGenerateStreamChunks(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FixedResponse = "Hello world test"
	sim := New(cfg)

	req := model.ChatCompletionRequest{
		Model: "test-model",
		Messages: []model.Message{
			{Role: "user", Content: "Hi"},
		},
		Stream: true,
	}

	chunks := sim.GenerateStreamChunks(req)

	// Should have: 1 role chunk + 3 word chunks + 1 final chunk = 5
	if len(chunks) != 5 {
		t.Fatalf("expected 5 chunks, got %d", len(chunks))
	}

	// First chunk: role
	if chunks[0].Choices[0].Delta == nil || chunks[0].Choices[0].Delta.Role != "assistant" {
		t.Error("first chunk should have role 'assistant'")
	}
	if chunks[0].Object != "chat.completion.chunk" {
		t.Errorf("expected object 'chat.completion.chunk', got %q", chunks[0].Object)
	}

	// Middle chunks: content words
	if chunks[1].Choices[0].Delta.Content != "Hello " {
		t.Errorf("expected 'Hello ', got %q", chunks[1].Choices[0].Delta.Content)
	}
	if chunks[2].Choices[0].Delta.Content != "world " {
		t.Errorf("expected 'world ', got %q", chunks[2].Choices[0].Delta.Content)
	}
	if chunks[3].Choices[0].Delta.Content != "test " {
		t.Errorf("expected 'test ', got %q", chunks[3].Choices[0].Delta.Content)
	}

	// Last chunk: finish_reason
	lastChunk := chunks[len(chunks)-1]
	if lastChunk.Choices[0].FinishReason == nil || *lastChunk.Choices[0].FinishReason != "stop" {
		t.Error("last chunk should have finish_reason 'stop'")
	}
}

func TestGetModels(t *testing.T) {
	cfg := DefaultConfig()
	sim := New(cfg)

	models := sim.GetModels()

	if models.Object != "list" {
		t.Errorf("expected object 'list', got %q", models.Object)
	}
	if len(models.Data) != len(cfg.AvailableModels) {
		t.Fatalf("expected %d models, got %d", len(cfg.AvailableModels), len(models.Data))
	}
	for i, m := range models.Data {
		if m.ID != cfg.AvailableModels[i] {
			t.Errorf("model %d: expected %q, got %q", i, cfg.AvailableModels[i], m.ID)
		}
		if m.Object != "model" {
			t.Errorf("model %d: expected object 'model', got %q", i, m.Object)
		}
	}
}

func TestEstimateStringTokens(t *testing.T) {
	tests := []struct {
		input    string
		minToken int
	}{
		{"", 0},
		{"Hi", 1},
		{"Hello, world! This is a test.", 7},
	}
	for _, tt := range tests {
		tokens := estimateStringTokens(tt.input)
		if tokens < tt.minToken {
			t.Errorf("estimateStringTokens(%q) = %d, want >= %d", tt.input, tokens, tt.minToken)
		}
	}
}
