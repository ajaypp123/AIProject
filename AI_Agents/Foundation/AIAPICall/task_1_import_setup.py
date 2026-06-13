#!/usr/bin/env python3
"""
Task 1: Import Required Libraries
Learn what libraries we need for AI API calls.
"""

# Step 1: Import the OpenAI library
# This library helps us talk to AI models
import openai  # TODO: Import "openai"

# Step 2: Import os for environment variables
# This helps us access API keys safely
import os  # TODO: Import "os"

print("✅ Step 1 Complete: Libraries imported!")
print("- openai: For making API calls")
print("- os: For accessing environment variables")

# Create marker
import os
os.makedirs("/tmp/markers", exist_ok=True)
with open("/tmp/markers/task1_imports_complete.txt", "w") as f:
    f.write("SUCCESS")