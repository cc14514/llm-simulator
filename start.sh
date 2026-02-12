#!/bin/bash
# Quick start script for LLM Simulator

set -e

echo "LLM Behavior Simulator - Quick Start"
echo "====================================="
echo ""

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    echo "Error: Python 3 is not installed. Please install Python 3.8 or higher."
    exit 1
fi

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "Activating virtual environment..."
source venv/bin/activate

# Install dependencies
echo "Installing dependencies..."
pip install -q -r requirements.txt

# Start the simulator
echo ""
echo "Starting LLM Behavior Simulator..."
echo "Server will be available at http://localhost:8000"
echo "Press Ctrl+C to stop"
echo ""

python simulator.py
