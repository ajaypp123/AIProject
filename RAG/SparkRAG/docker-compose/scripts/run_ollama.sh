#!/usr/bin/env bash

# Define target functions
start_service() {
    rm -rf /scripts/ollama_initalized # Remove the initalized file to ensure Ollama starts fresh

    echo "Starting Ollama server..."
    ollama serve &  # Start Ollama in the background
    sleep 30

    echo "Ollama is ready, installing the model..."

    ollama pull gemma3:4b
    ollama pull nomic-embed-text
    ollama list

    touch /scripts/ollama_initalized # Create the initalized file to indicate Ollama has started
    echo "Ollama is ready, with model...."

    sleep infinity  # Keep the script running to keep Ollama alive
}

stop_service() {
    echo "Stopping service..."
}

check_health() {
    if [ -f /scripts/ollama_initalized ]; then
        echo "Ollama is healthy."
        exit 0
    else
        echo "Ollama is not healthy."
        exit 1
    fi
}

# Switch / Dispatch pattern
case "$1" in
    start)
        shift # Remove "start" so "$@" contains only the subsequent arguments
        start_service "$@"
        ;;
    stop)
        stop_service
        ;;
    healthz)
        check_health
        ;;
    *)
        echo "Usage: $0 {start|stop|healthz} [arguments]"
        exit 1
        ;;
esac
