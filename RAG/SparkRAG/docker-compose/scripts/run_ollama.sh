#!/bin/bash

echo "Starting Ollama server..."
ollama serve &  # Start Ollama in the background
sleep 30

echo "Ollama is ready, installing the model..."

ollama pull gemma3:4b
ollama pull nomic-embed-text
ollama list

ollama run gemma3:4b
ollama run nomic-embed-text

echo "Ollama is ready, with model...."

sleep infinity  # Keep the script running to keep Ollama alive
