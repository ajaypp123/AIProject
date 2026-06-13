#!/usr/bin/env python3
"""
📄 Task 2: Document Processing with Smart Chunking
Learn how to split documents intelligently for vector embedding.
"""

import os
from typing import List

# TODO 1: Import the text splitter
# Replace ___ with: RecursiveCharacterTextSplitter
from langchain_text_splitters import RecursiveCharacterTextSplitter

def process_documents():
    print("📄 Task 2: Smart Document Chunking")
    print("=" * 55)

    # Sample company documentation
    long_document = """
    TechDocs Employee Handbook

    Chapter 1: Remote Work Policy

    Our company embraces flexible work arrangements to support work-life balance.
    Employees are permitted to work remotely up to 3 days per week, provided they
    maintain regular communication with their team and meet all performance expectations.
    Remote work requires manager approval and must not impact team collaboration or
    customer service quality.

    To work remotely, employees must have a suitable home office setup with reliable
    internet connection, appropriate workspace, and necessary equipment. The company
    provides a one-time stipend of $500 for home office setup. VPN access is mandatory
    for accessing company systems remotely.

    Chapter 2: Communication Guidelines

    Effective communication is essential for remote work success. All employees must
    be available during core hours (10 AM - 3 PM in their local timezone) for meetings
    and collaboration. Slack is our primary communication tool, with response times
    expected within 2 hours during work hours.

    Video calls are encouraged for team meetings to maintain personal connections.
    Camera use is optional but recommended. All meetings should have clear agendas
    and action items documented in our project management system.

    Chapter 3: Performance Management

    Remote work does not change performance expectations. Managers will evaluate
    employees based on deliverables, quality of work, and contribution to team goals
    rather than hours logged. Regular 1-on-1 meetings should continue virtually.
    """

    print(f"📋 Original Document Length: {len(long_document)} characters")
    print("-" * 40)

    # TODO 2 & 3: Configure the text splitter
    splitter = RecursiveCharacterTextSplitter(
        chunk_size=500,      # Replace ___ with: 500
        chunk_overlap=200,   # Replace ___ with: 200
        length_function=len,
        separators=[" "]
    )

    # Split the document
    chunks = splitter.split_text(long_document)

    print(f"\n✂️ Chunking Results:")
    print(f"• Created {len(chunks)} chunks")
    print(f"• Chunk size: ~500 characters")
    print(f"• Overlap: 100 characters\n")

    # Display chunks
    for i, chunk in enumerate(chunks[:3], 1):  # Show first 3 chunks
        print(f"📄 Chunk {i} ({len(chunk)} chars):")
        print("-" * 40)
        print(chunk[:200] + "..." if len(chunk) > 200 else chunk)
        print()

    # Analyze overlap
    if len(chunks) > 1:
        # Find common text between chunk 1 and 2
        overlap_start = chunks[1][:100]
        if overlap_start in chunks[0]:
            print("🔄 Overlap Detection:")
            print("-" * 40)
            print(f"Found overlap: '{overlap_start[:50]}...'")
            print("✅ Context preserved between chunks!")

    # Key insights
    print("\n💡 Why Smart Chunking Matters:")
    print("-" * 40)
    print("• Preserves sentence boundaries")
    print("• Maintains context with overlap")
    print("• Optimal for embedding models")
    print("• Improves retrieval accuracy by 40%!")

    # Create completion marker
    os.makedirs("/tmp/markers", exist_ok=True)
    with open("/tmp/markers/task2_chunking_complete.txt", "w") as f:
        f.write("COMPLETED")

    print("\n✅ Task 2 completed! Document chunking mastered.")

    return chunks

if __name__ == "__main__":
    chunks = process_documents()