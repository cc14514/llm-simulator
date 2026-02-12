#!/usr/bin/env python3
"""
Example usage of the LLM Behavior Simulator with OpenAI Python client

Install the OpenAI client first:
    pip install openai

Then run this script while the simulator is running:
    python simulator.py  # In one terminal
    python example_usage.py  # In another terminal
"""

from openai import OpenAI


def example_basic_completion():
    """Example of basic chat completion"""
    print("=" * 60)
    print("Example 1: Basic Chat Completion")
    print("=" * 60)
    
    client = OpenAI(
        api_key="not-needed",  # Simulator doesn't validate API keys
        base_url="http://localhost:8000/v1"
    )
    
    response = client.chat.completions.create(
        model="gpt-3.5-turbo",
        messages=[
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "What is the capital of France?"}
        ]
    )
    
    print(f"\nModel: {response.model}")
    print(f"Response: {response.choices[0].message.content}")
    print(f"Tokens used: {response.usage.total_tokens}")
    print()


def example_streaming_completion():
    """Example of streaming chat completion"""
    print("=" * 60)
    print("Example 2: Streaming Chat Completion")
    print("=" * 60)
    
    client = OpenAI(
        api_key="not-needed",
        base_url="http://localhost:8000/v1"
    )
    
    print("\nStreaming response: ", end="", flush=True)
    
    stream = client.chat.completions.create(
        model="gpt-4-turbo",
        messages=[
            {"role": "user", "content": "Tell me a short story"}
        ],
        stream=True
    )
    
    for chunk in stream:
        if chunk.choices[0].delta.content:
            print(chunk.choices[0].delta.content, end="", flush=True)
    
    print("\n")


def example_list_models():
    """Example of listing available models"""
    print("=" * 60)
    print("Example 3: List Available Models")
    print("=" * 60)
    
    client = OpenAI(
        api_key="not-needed",
        base_url="http://localhost:8000/v1"
    )
    
    models = client.models.list()
    
    print("\nAvailable models:")
    for model in models.data:
        print(f"  - {model.id}")
    print()


def example_with_parameters():
    """Example with various parameters"""
    print("=" * 60)
    print("Example 4: Chat Completion with Parameters")
    print("=" * 60)
    
    client = OpenAI(
        api_key="not-needed",
        base_url="http://localhost:8000/v1"
    )
    
    response = client.chat.completions.create(
        model="gpt-4o",
        messages=[
            {"role": "user", "content": "Explain quantum computing in simple terms"}
        ],
        temperature=0.7,
        max_tokens=100,
        top_p=0.9
    )
    
    print(f"\nModel: {response.model}")
    print(f"Response: {response.choices[0].message.content}")
    print(f"Finish reason: {response.choices[0].finish_reason}")
    print()


def main():
    """Run all examples"""
    print("\nLLM Behavior Simulator - Usage Examples")
    print("Make sure the simulator is running on localhost:8000\n")
    
    try:
        example_basic_completion()
        example_streaming_completion()
        example_list_models()
        example_with_parameters()
        
        print("=" * 60)
        print("All examples completed successfully!")
        print("=" * 60)
        
    except Exception as e:
        print(f"\nError: {e}")
        print("\nMake sure the simulator is running:")
        print("  python simulator.py")


if __name__ == "__main__":
    main()
