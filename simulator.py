#!/usr/bin/env python3
"""
LLM Behavior Simulator
A minimal OpenAI-compatible API server for testing AI gateways.
"""

import asyncio
import json
import time
import uuid
from typing import List, Optional, Dict, Any

from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field
import uvicorn


# Request/Response Models
class Message(BaseModel):
    role: str
    content: str


class ChatCompletionRequest(BaseModel):
    model: str = "gpt-3.5-turbo"
    messages: List[Message]
    temperature: Optional[float] = 1.0
    max_tokens: Optional[int] = None
    stream: Optional[bool] = False
    top_p: Optional[float] = 1.0
    n: Optional[int] = 1
    stop: Optional[List[str]] = None


class Usage(BaseModel):
    prompt_tokens: int
    completion_tokens: int
    total_tokens: int


class Choice(BaseModel):
    index: int
    message: Message
    finish_reason: str


class ChatCompletionResponse(BaseModel):
    id: str
    object: str = "chat.completion"
    created: int
    model: str
    choices: List[Choice]
    usage: Usage


class Model(BaseModel):
    id: str
    object: str = "model"
    created: int
    owned_by: str = "simulator"


class ModelList(BaseModel):
    object: str = "list"
    data: List[Model]


# Create FastAPI app
app = FastAPI(
    title="LLM Behavior Simulator",
    description="A minimal OpenAI-compatible API server for testing AI gateways",
    version="1.0.0"
)


# Available models
AVAILABLE_MODELS = [
    "gpt-3.5-turbo",
    "gpt-4",
    "gpt-4-turbo",
    "gpt-4o",
    "gpt-4o-mini",
]


def generate_response_text(messages: List[Message], model: str) -> str:
    """
    Generate a simple response based on the input messages.
    This is a minimal simulator, so we just echo back information about the request.
    """
    last_message = messages[-1].content if messages else "No message"
    truncated_message = last_message[:50]
    ellipsis = "..." if len(last_message) > 50 else ""
    response = f"[Simulator Response] Model: {model}, Message received: '{truncated_message}{ellipsis}'"
    return response


def estimate_tokens(text: str) -> int:
    """Simple token estimation (roughly 4 characters per token)"""
    return len(text) // 4


@app.get("/")
async def root():
    """Root endpoint with API information"""
    return {
        "name": "LLM Behavior Simulator",
        "version": "1.0.0",
        "description": "OpenAI-compatible API for testing AI gateways",
        "endpoints": [
            "/v1/chat/completions",
            "/v1/models"
        ]
    }


@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "healthy"}


@app.get("/v1/models")
async def list_models() -> ModelList:
    """List available models"""
    models = [
        Model(
            id=model_id,
            created=int(time.time()),
            owned_by="simulator"
        )
        for model_id in AVAILABLE_MODELS
    ]
    return ModelList(data=models)


async def generate_stream(request: ChatCompletionRequest):
    """Generate streaming response"""
    response_text = generate_response_text(request.messages, request.model)
    request_id = f"chatcmpl-{uuid.uuid4().hex[:24]}"
    created = int(time.time())
    
    # Split response into chunks
    words = response_text.split()
    
    for i, word in enumerate(words):
        chunk = {
            "id": request_id,
            "object": "chat.completion.chunk",
            "created": created,
            "model": request.model,
            "choices": [
                {
                    "index": 0,
                    "delta": {"content": word + " "} if i > 0 else {"role": "assistant", "content": word + " "},
                    "finish_reason": None
                }
            ]
        }
        yield f"data: {json.dumps(chunk)}\n\n"
        await asyncio.sleep(0.05)  # Simulate processing delay
    
    # Send final chunk
    final_chunk = {
        "id": request_id,
        "object": "chat.completion.chunk",
        "created": created,
        "model": request.model,
        "choices": [
            {
                "index": 0,
                "delta": {},
                "finish_reason": "stop"
            }
        ]
    }
    yield f"data: {json.dumps(final_chunk)}\n\n"
    yield "data: [DONE]\n\n"


@app.post("/v1/chat/completions")
async def create_chat_completion(request: ChatCompletionRequest):
    """Create a chat completion"""
    
    # Validate model
    if request.model not in AVAILABLE_MODELS:
        raise HTTPException(
            status_code=400,
            detail=f"Model {request.model} not found. Available models: {AVAILABLE_MODELS}"
        )
    
    # Handle streaming
    if request.stream:
        return StreamingResponse(
            generate_stream(request),
            media_type="text/event-stream"
        )
    
    # Non-streaming response
    response_text = generate_response_text(request.messages, request.model)
    
    # Calculate token usage
    prompt_text = " ".join([msg.content for msg in request.messages])
    prompt_tokens = estimate_tokens(prompt_text)
    completion_tokens = estimate_tokens(response_text)
    
    response = ChatCompletionResponse(
        id=f"chatcmpl-{uuid.uuid4().hex[:24]}",
        created=int(time.time()),
        model=request.model,
        choices=[
            Choice(
                index=0,
                message=Message(role="assistant", content=response_text),
                finish_reason="stop"
            )
        ],
        usage=Usage(
            prompt_tokens=prompt_tokens,
            completion_tokens=completion_tokens,
            total_tokens=prompt_tokens + completion_tokens
        )
    )
    
    return response


def main():
    """Run the simulator server"""
    import argparse
    
    parser = argparse.ArgumentParser(description="LLM Behavior Simulator")
    parser.add_argument("--host", default="0.0.0.0", help="Host to bind to")
    parser.add_argument("--port", type=int, default=8000, help="Port to bind to")
    parser.add_argument("--reload", action="store_true", help="Enable auto-reload")
    
    args = parser.parse_args()
    
    print(f"Starting LLM Behavior Simulator on {args.host}:{args.port}")
    print(f"OpenAI-compatible API available at http://{args.host}:{args.port}/v1")
    
    uvicorn.run(
        "simulator:app",
        host=args.host,
        port=args.port,
        reload=args.reload
    )


if __name__ == "__main__":
    main()
