# LLM Behavior Simulator

A minimal LLM behavior simulator for validating AI gateway functionality.

## Overview

This simulator provides an OpenAI-compatible API server that mimics the behavior of LLM endpoints. It's designed for testing and validating AI gateway functionality, including:

- Request routing and proxying
- Authentication and authorization flows
- Rate limiting and throttling
- Load balancing
- Request/response transformation
- Monitoring and observability

## Features

- ✅ OpenAI-compatible `/v1/chat/completions` endpoint
- ✅ Support for both streaming and non-streaming responses
- ✅ Model listing via `/v1/models` endpoint
- ✅ Health check endpoint
- ✅ Simple token usage estimation
- ✅ Lightweight and easy to deploy
- ✅ No external dependencies on actual LLM providers

## Installation

### Requirements

- Python 3.8 or higher

### Setup

1. Clone the repository:
```bash
git clone https://github.com/cc14514/llm-simulator.git
cd llm-simulator
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

### Docker Deployment

You can also run the simulator using Docker:

```bash
# Build and run with Docker Compose
docker-compose up -d

# Or build and run manually
docker build -t llm-simulator .
docker run -p 8000:8000 llm-simulator
```

### Publish to GHCR (multi-arch)

This repo includes a `Makefile` target that builds and pushes a multi-architecture image for `linux/amd64` and `linux/arm64` using Docker Buildx.

1. Login to GHCR (use a GitHub token with `write:packages`):

```bash
echo "$GITHUB_TOKEN" | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
```

2. Build & push:

```bash
# Pushes: ghcr.io/your-org/llm-simulator:latest
make docker-push REGISTRY=ghcr.io/your-org

# Optional: set a tag
make docker-push REGISTRY=ghcr.io/your-org TAG=v0.1.0
```

## Usage

### Quick Start

The easiest way to get started:

```bash
./start.sh
```

This script will:
- Create a virtual environment
- Install dependencies
- Start the simulator on localhost:8000

### Starting the Server

Run the simulator with default settings (localhost:8000):

```bash
python simulator.py
```

Custom host and port:

```bash
python simulator.py --host 0.0.0.0 --port 8080
```

With auto-reload for development:

```bash
python simulator.py --reload
```

### Available Endpoints

- `GET /` - API information
- `GET /health` - Health check
- `GET /v1/models` - List available models
- `POST /v1/chat/completions` - Create chat completion

### Example Requests

#### Non-streaming Chat Completion

```bash
curl http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

#### Streaming Chat Completion

```bash
curl http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Tell me a story"}
    ],
    "stream": true
  }'
```

#### List Available Models

```bash
curl http://localhost:8000/v1/models
```

#### Health Check

```bash
curl http://localhost:8000/health
```

### Using with OpenAI Client Libraries

The simulator is compatible with OpenAI client libraries. Just point the base URL to your simulator instance:

#### Python Example

```python
from openai import OpenAI

client = OpenAI(
    api_key="dummy-key",  # Simulator doesn't validate API keys
    base_url="http://localhost:8000/v1"
)

response = client.chat.completions.create(
    model="gpt-3.5-turbo",
    messages=[
        {"role": "user", "content": "Hello!"}
    ]
)

print(response.choices[0].message.content)
```

#### Streaming Example

```python
from openai import OpenAI

client = OpenAI(
    api_key="dummy-key",
    base_url="http://localhost:8000/v1"
)

stream = client.chat.completions.create(
    model="gpt-3.5-turbo",
    messages=[{"role": "user", "content": "Count to 10"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

## Supported Models

The simulator supports the following model identifiers:

- `gpt-3.5-turbo`
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-4o-mini`

## Use Cases

### AI Gateway Testing

Use this simulator to test your AI gateway without incurring costs from actual LLM providers:

1. **Routing Logic**: Verify request routing to different backends
2. **Authentication**: Test API key validation and authorization
3. **Rate Limiting**: Validate rate limiting and throttling mechanisms
4. **Load Balancing**: Test load distribution across multiple backends
5. **Caching**: Verify response caching behavior
6. **Monitoring**: Test logging, metrics, and tracing integration

### Development and CI/CD

- Use in development environments without API costs
- Include in CI/CD pipelines for integration testing
- Mock LLM responses for testing downstream applications

## Configuration

The simulator can be configured via command-line arguments:

| Argument | Default | Description |
|----------|---------|-------------|
| `--host` | `0.0.0.0` | Host address to bind to |
| `--port` | `8000` | Port number to listen on |
| `--reload` | `false` | Enable auto-reload for development |

## Architecture

The simulator is built with:

- **FastAPI**: Modern, fast web framework for building APIs
- **Pydantic**: Data validation using Python type annotations
- **Uvicorn**: ASGI server for running the application
- **SSE-Starlette**: Server-Sent Events support for streaming

## Limitations

This is a minimal simulator designed for testing purposes:

- Responses are simple echo messages, not actual AI-generated content
- No authentication or authorization (accepts any API key)
- Token counting is approximate (4 characters ≈ 1 token)
- No persistent state or history
- Limited error handling for edge cases

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
