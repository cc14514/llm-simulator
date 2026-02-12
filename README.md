# llm-simulator

A minimal LLM behavior simulator for validating AI gateway functionality. It exposes OpenAI-compatible API endpoints that return deterministic, configurable responses — useful for testing AI gateways, proxies, and integrations without calling real LLM providers.

## Features

- **OpenAI-compatible API** — `/v1/chat/completions` and `/v1/models` endpoints
- **Streaming support** — Server-Sent Events (SSE) streaming responses
- **Echo mode** — Echoes back the last user message for request validation
- **Configurable latency** — Simulate response delays and per-chunk streaming delays
- **Error injection** — Configurable error rate and status codes for resilience testing
- **Token estimation** — Approximate token counts in responses

## Quick Start

```bash
go build -o llm-simulator .
./llm-simulator
```

The server starts on port `8080` by default.

## Usage

### Non-streaming request

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

### Streaming request

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

### List models

```bash
curl http://localhost:8080/v1/models
```

### Health check

```bash
curl http://localhost:8080/health
```

## Configuration

All settings are configurable via command-line flags:

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `8080` | Port to listen on |
| `-response-delay` | `0` | Artificial delay before responding (e.g. `100ms`, `1s`) |
| `-stream-chunk-delay` | `50ms` | Delay between SSE stream chunks |
| `-echo` | `false` | Echo back the last user message |
| `-response` | `This is a simulated response from the LLM simulator.` | Fixed response text |
| `-error-rate` | `0` | Probability (0.0–1.0) of returning a simulated error |
| `-error-status` | `500` | HTTP status code for simulated errors |
| `-models` | `llm-simulator-1,gpt-4o,gpt-4o-mini` | Comma-separated list of available models |

The port can also be set via the `LLM_SIM_PORT` environment variable.

### Example: Echo mode with latency

```bash
./llm-simulator -echo -response-delay 200ms -stream-chunk-delay 100ms
```

### Example: Error injection testing

```bash
./llm-simulator -error-rate 0.5 -error-status 503
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/v1/models` | List available models |
| `POST` | `/v1/chat/completions` | Chat completion (streaming and non-streaming) |

## Running Tests

```bash
go test ./...
```

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
