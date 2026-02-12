package simulator

import (
	"fmt"
	"strings"
	"time"

	"github.com/cc14514/llm-simulator/internal/model"
)

// Config controls simulator behavior.
type Config struct {
	// DefaultModel is the model name returned in responses.
	DefaultModel string
	// AvailableModels lists models reported by the /v1/models endpoint.
	AvailableModels []string
	// ResponseDelay adds artificial latency before responding.
	ResponseDelay time.Duration
	// StreamChunkDelay adds delay between SSE chunks during streaming.
	StreamChunkDelay time.Duration
	// EchoMode when true echoes back the last user message content.
	EchoMode bool
	// FixedResponse is the fixed response text when EchoMode is false.
	FixedResponse string
	// ErrorRate is the probability (0.0–1.0) of returning a simulated error.
	ErrorRate float64
	// ErrorStatusCode is the HTTP status code used for simulated errors.
	ErrorStatusCode int
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		DefaultModel: "llm-simulator-1",
		AvailableModels: []string{
			"llm-simulator-1",
			"gpt-4o",
			"gpt-4o-mini",
		},
		ResponseDelay:    0,
		StreamChunkDelay: 50 * time.Millisecond,
		EchoMode:         false,
		FixedResponse:    "This is a simulated response from the LLM simulator.",
		ErrorRate:        0,
		ErrorStatusCode:  500,
	}
}

// Simulator generates deterministic LLM responses.
type Simulator struct {
	Config Config
}

// New creates a new Simulator with the given config.
func New(cfg Config) *Simulator {
	return &Simulator{Config: cfg}
}

// GenerateResponse produces a simulated chat completion response.
func (s *Simulator) GenerateResponse(req model.ChatCompletionRequest) model.ChatCompletionResponse {
	responseText := s.getResponseText(req)
	finishReason := "stop"
	resolvedModel := req.Model
	if resolvedModel == "" {
		resolvedModel = s.Config.DefaultModel
	}

	promptTokens := s.estimateTokens(req.Messages)
	completionTokens := estimateStringTokens(responseText)

	return model.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-sim-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resolvedModel,
		Choices: []model.Choice{
			{
				Index: 0,
				Message: &model.Message{
					Role:    "assistant",
					Content: responseText,
				},
				FinishReason: &finishReason,
			},
		},
		Usage: model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
}

// GenerateStreamChunks produces a sequence of SSE-compatible streaming chunks.
func (s *Simulator) GenerateStreamChunks(req model.ChatCompletionRequest) []model.ChatCompletionResponse {
	responseText := s.getResponseText(req)
	resolvedModel := req.Model
	if resolvedModel == "" {
		resolvedModel = s.Config.DefaultModel
	}
	id := fmt.Sprintf("chatcmpl-sim-%d", time.Now().UnixNano())

	words := strings.Fields(responseText)
	chunks := make([]model.ChatCompletionResponse, 0, len(words)+1)

	// Role chunk
	chunks = append(chunks, model.ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   resolvedModel,
		Choices: []model.Choice{
			{
				Index: 0,
				Delta: &model.Message{
					Role: "assistant",
				},
				FinishReason: nil,
			},
		},
	})

	// Content chunks (word by word)
	for _, word := range words {
		chunks = append(chunks, model.ChatCompletionResponse{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   resolvedModel,
			Choices: []model.Choice{
				{
					Index: 0,
					Delta: &model.Message{
						Content: word + " ",
					},
					FinishReason: nil,
				},
			},
		})
	}

	// Final chunk with finish_reason
	finishReason := "stop"
	chunks = append(chunks, model.ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   resolvedModel,
		Choices: []model.Choice{
			{
				Index:        0,
				Delta:        &model.Message{},
				FinishReason: &finishReason,
			},
		},
	})

	return chunks
}

// GetModels returns the list of available models.
func (s *Simulator) GetModels() model.ModelList {
	data := make([]model.ModelInfo, len(s.Config.AvailableModels))
	for i, m := range s.Config.AvailableModels {
		data[i] = model.ModelInfo{
			ID:      m,
			Object:  "model",
			Created: 1700000000,
			OwnedBy: "llm-simulator",
		}
	}
	return model.ModelList{
		Object: "list",
		Data:   data,
	}
}

func (s *Simulator) getResponseText(req model.ChatCompletionRequest) string {
	if s.Config.EchoMode && len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		return fmt.Sprintf("Echo: %s", lastMsg.Content)
	}
	return s.Config.FixedResponse
}

func (s *Simulator) estimateTokens(messages []model.Message) int {
	total := 0
	for _, m := range messages {
		total += estimateStringTokens(m.Content)
		total += 4 // overhead per message (role, etc.)
	}
	return total
}

func estimateStringTokens(s string) int {
	// Rough approximation: ~4 chars per token
	tokens := len(s) / 4
	if tokens == 0 && len(s) > 0 {
		tokens = 1
	}
	return tokens
}
