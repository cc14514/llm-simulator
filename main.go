package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cc14514/llm-simulator/internal/handler"
	"github.com/cc14514/llm-simulator/internal/simulator"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	responseDelay := flag.Duration("response-delay", 0, "artificial delay before responding (e.g. 100ms, 1s)")
	streamChunkDelay := flag.Duration("stream-chunk-delay", 50*time.Millisecond, "delay between stream chunks")
	echoMode := flag.Bool("echo", false, "echo back the last user message")
	fixedResponse := flag.String("response", "This is a simulated response from the LLM simulator.", "fixed response text")
	errorRate := flag.Float64("error-rate", 0, "probability of returning a simulated error (0.0-1.0)")
	errorStatusCode := flag.Int("error-status", 500, "HTTP status code for simulated errors")
	models := flag.String("models", "llm-simulator-1,gpt-4o,gpt-4o-mini", "comma-separated list of available models")
	flag.Parse()

	cfg := simulator.Config{
		DefaultModel:     "llm-simulator-1",
		AvailableModels:  splitModels(*models),
		ResponseDelay:    *responseDelay,
		StreamChunkDelay: *streamChunkDelay,
		EchoMode:         *echoMode,
		FixedResponse:    *fixedResponse,
		ErrorRate:        *errorRate,
		ErrorStatusCode:  *errorStatusCode,
	}

	if envPort := os.Getenv("LLM_SIM_PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", port)
	}

	sim := simulator.New(cfg)
	h := handler.New(sim)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/v1/models", h.ListModels)
	mux.HandleFunc("/v1/chat/completions", h.ChatCompletions)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("LLM Simulator listening on %s", addr)
	log.Printf("  Echo mode: %v", cfg.EchoMode)
	log.Printf("  Response delay: %v", cfg.ResponseDelay)
	log.Printf("  Stream chunk delay: %v", cfg.StreamChunkDelay)
	log.Printf("  Error rate: %.2f", cfg.ErrorRate)
	log.Printf("  Models: %v", cfg.AvailableModels)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func splitModels(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
