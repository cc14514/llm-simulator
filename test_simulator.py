#!/usr/bin/env python3
"""
Test script for the LLM Behavior Simulator

This script tests the basic functionality of the simulator endpoints.
"""

import requests
import json
import sys
import time


def test_health(base_url):
    """Test health endpoint"""
    print("Testing health endpoint...")
    response = requests.get(f"{base_url}/health")
    assert response.status_code == 200, f"Health check failed: {response.status_code}"
    data = response.json()
    assert data["status"] == "healthy", f"Unexpected health status: {data}"
    print("✓ Health endpoint working")
    return True


def test_root(base_url):
    """Test root endpoint"""
    print("\nTesting root endpoint...")
    response = requests.get(base_url)
    assert response.status_code == 200, f"Root endpoint failed: {response.status_code}"
    data = response.json()
    assert "name" in data, "Missing 'name' in root response"
    assert "endpoints" in data, "Missing 'endpoints' in root response"
    print(f"✓ Root endpoint working: {data['name']}")
    return True


def test_list_models(base_url):
    """Test models listing"""
    print("\nTesting models endpoint...")
    response = requests.get(f"{base_url}/v1/models")
    assert response.status_code == 200, f"Models endpoint failed: {response.status_code}"
    data = response.json()
    assert "data" in data, "Missing 'data' in models response"
    assert len(data["data"]) > 0, "No models returned"
    print(f"✓ Models endpoint working: {len(data['data'])} models available")
    for model in data["data"][:3]:
        print(f"  - {model['id']}")
    return True


def test_chat_completion(base_url):
    """Test non-streaming chat completion"""
    print("\nTesting chat completion (non-streaming)...")
    payload = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {"role": "user", "content": "Hello, this is a test message!"}
        ]
    }
    response = requests.post(
        f"{base_url}/v1/chat/completions",
        json=payload,
        headers={"Content-Type": "application/json"}
    )
    assert response.status_code == 200, f"Chat completion failed: {response.status_code}"
    data = response.json()
    
    # Validate response structure
    assert "id" in data, "Missing 'id' in response"
    assert "choices" in data, "Missing 'choices' in response"
    assert len(data["choices"]) > 0, "No choices in response"
    assert "message" in data["choices"][0], "Missing 'message' in choice"
    assert "content" in data["choices"][0]["message"], "Missing 'content' in message"
    assert "usage" in data, "Missing 'usage' in response"
    
    print("✓ Chat completion working")
    print(f"  Response: {data['choices'][0]['message']['content'][:80]}...")
    print(f"  Tokens: {data['usage']['total_tokens']}")
    return True


def test_chat_completion_streaming(base_url):
    """Test streaming chat completion"""
    print("\nTesting chat completion (streaming)...")
    payload = {
        "model": "gpt-4",
        "messages": [
            {"role": "user", "content": "Count to five"}
        ],
        "stream": True
    }
    response = requests.post(
        f"{base_url}/v1/chat/completions",
        json=payload,
        headers={"Content-Type": "application/json"},
        stream=True
    )
    assert response.status_code == 200, f"Streaming chat completion failed: {response.status_code}"
    
    chunks_received = 0
    content_pieces = []
    
    for line in response.iter_lines():
        if line:
            line = line.decode('utf-8')
            if line.startswith('data: '):
                data_str = line[6:]
                if data_str == '[DONE]':
                    break
                try:
                    chunk = json.loads(data_str)
                    chunks_received += 1
                    if 'choices' in chunk and len(chunk['choices']) > 0:
                        delta = chunk['choices'][0].get('delta', {})
                        if 'content' in delta:
                            content_pieces.append(delta['content'])
                except json.JSONDecodeError:
                    pass
    
    assert chunks_received > 0, "No chunks received in streaming response"
    print(f"✓ Streaming chat completion working")
    print(f"  Chunks received: {chunks_received}")
    print(f"  Content: {''.join(content_pieces)[:80]}...")
    return True


def test_invalid_model(base_url):
    """Test with invalid model"""
    print("\nTesting invalid model handling...")
    payload = {
        "model": "invalid-model-xyz",
        "messages": [
            {"role": "user", "content": "Hello"}
        ]
    }
    response = requests.post(
        f"{base_url}/v1/chat/completions",
        json=payload,
        headers={"Content-Type": "application/json"}
    )
    assert response.status_code == 400, f"Expected 400 for invalid model, got: {response.status_code}"
    print("✓ Invalid model properly rejected")
    return True


def main():
    """Run all tests"""
    base_url = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8000"
    
    print(f"Testing LLM Behavior Simulator at {base_url}")
    print("=" * 60)
    
    # Give server a moment to be ready
    time.sleep(1)
    
    tests = [
        test_health,
        test_root,
        test_list_models,
        test_chat_completion,
        test_chat_completion_streaming,
        test_invalid_model,
    ]
    
    passed = 0
    failed = 0
    
    for test in tests:
        try:
            if test(base_url):
                passed += 1
        except Exception as e:
            print(f"✗ Test failed: {e}")
            failed += 1
    
    print("\n" + "=" * 60)
    print(f"Tests completed: {passed} passed, {failed} failed")
    
    if failed > 0:
        sys.exit(1)
    else:
        print("\n✓ All tests passed!")
        sys.exit(0)


if __name__ == "__main__":
    main()
