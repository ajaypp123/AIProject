#!/usr/bin/env bash

echo "Starting Ingestor..."

echo "Checking version of rag-ingestor_cli.py..."
/app/rag-ingestor_cli.py --version

echo "Wait till the Ingestor is healthy..."
while ! /app/rag-ingestor_cli.py health; do
    echo "Ingestor is not healthy yet. Waiting..."
    sleep 5
done
echo "Ingestor is healthy..."

echo "Cleaning up the old Ingestor rag..."
if [ -f "/data/ingestor/.ingestor.json" ]; then
    /app/rag-ingestor_cli.py delete all
    rm -rf /data/ingestor/.ingestor.json
else
    echo "Old data in RAG not found. Skipping cleanup."
fi


echo "Starting the Ingestor..."
/app/rag-ingestor_cli.py $INGESTOR_OPERATION_MODE

tail -f /dev/null