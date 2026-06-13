#!/usr/bin/env python3
"""
Task 1: Environment & Vector Store Setup
Bridge your Vector Databases Lab knowledge to create the foundation for RAG
"""

import os
import chromadb
from sentence_transformers import SentenceTransformer

print("🔧 Task 1: Setting up Vector Store for RAG")
print("=" * 50)

# TODO 1: Initialize ChromaDB client for persistent storage
# Hint: Use chromadb.PersistentClient(path="./chroma_db")
client = chromadb.PersistentClient(path="___")  # Replace ___ with "./chroma_db"

print("✅ ChromaDB client initialized")

# TODO 2: Create or get collection named "techcorp_rag"
# Hint: Use client.get_or_create_collection(name="techcorp_rag")
collection = client.get_or_create_collection(name="___")  # Replace ___ with "techcorp_rag"

print(f"✅ Collection '{collection.name}' ready")

# TODO 3: Initialize embedding model for 384-dimension vectors
# Hint: Use SentenceTransformer("all-MiniLM-L6-v2")
model = SentenceTransformer("___")  # Replace ___ with "all-MiniLM-L6-v2"

print("✅ Embedding model loaded")

# Test the setup
test_text = "Testing RAG setup"
test_embedding = model.encode(test_text)
print(f"✅ Test embedding created: {len(test_embedding)} dimensions")

# Verify everything works
print("\n" + "=" * 50)
print("🎉 SUCCESS! Your vector store is ready for RAG!")
print(f"   - ChromaDB initialized")
print(f"   - Collection: {collection.name}")
print(f"   - Embedding model: all-MiniLM-L6-v2")
print(f"   - Vector dimensions: {len(test_embedding)}")
print("=" * 50)

# Create marker file
os.makedirs("/root/markers", exist_ok=True)
with open("/root/markers/task1_setup_complete.txt", "w") as f:
    f.write("TASK1_COMPLETE")

print("\n💡 Remember: You learned this in Vector Databases Lab - now applying it for RAG!")
print("\n✅ Task 1 completed!")