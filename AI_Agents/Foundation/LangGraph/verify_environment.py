#!/usr/bin/env python3
"""
Environment Verification Script for LangGraph Lab
Checks all required packages and configurations
"""

import os
import sys

print("🔧 Verifying LangGraph Lab Environment")
print("=" * 50)

# Check Python version
print(f"✓ Python version: {sys.version.split()[0]}")

# Check required packages
packages_to_check = {
    "langgraph": "Stateful graph framework",
    "langchain": "LLM orchestration",
    "langchain_openai": "OpenAI integration",
    "langchain_community": "Community tools",
    "duckduckgo_search": "Web search (FREE!)",
    "bs4": "Web scraping",
    "dotenv": "Environment management"
}

print("\n📦 Checking installed packages:")
missing_packages = []

for package, description in packages_to_check.items():
    try:
        if package == "langchain_openai":
            import langchain_openai
        elif package == "langchain_community":
            import langchain_community
        elif package == "duckduckgo_search":
            from duckduckgo_search import DDGS
        elif package == "bs4":
            from bs4 import BeautifulSoup
        elif package == "dotenv":
            from dotenv import load_dotenv
        else:
            __import__(package)
        print(f"  ✓ {package}: {description}")
    except ImportError:
        print(f"  ✗ {package}: NOT INSTALLED")
        missing_packages.append(package)

# Check environment variables
print("\n🔑 Checking environment variables:")
env_vars = ["OPENAI_API_KEY", "OPENAI_API_BASE"]
missing_env = []

for var in env_vars:
    if os.getenv(var):
        masked_value = os.getenv(var)[:8] + "..." if len(os.getenv(var)) > 8 else "***"
        print(f"  ✓ {var}: {masked_value}")
    else:
        print(f"  ✗ {var}: NOT SET")
        missing_env.append(var)

# Check model availability
print("\n🤖 Available models via proxy:")
models = [
    "openai/gpt-4.1-mini",
    "openai/gpt-5-mini",
    "deepseek/deepseek-chat",
    "x-ai/grok-code-fast-1"
]
for model in models:
    print(f"  • {model}")

# Final status
print("\n" + "=" * 50)
if missing_packages:
    print("❌ Missing packages detected!")
    print("Run the following command to install:")
    print(f"pip install {' '.join(missing_packages)}")
    sys.exit(1)
elif missing_env:
    print("⚠️  Missing environment variables!")
    print("Please set the following variables:")
    for var in missing_env:
        print(f"  export {var}='your_value_here'")
else:
    print("✅ Environment verification complete!")
    print("🎉 All requirements satisfied - ready for LangGraph!")

    # Create marker file
    os.makedirs("/root/markers", exist_ok=True)
    with open("/root/markers/environment_verified.txt", "w") as f:
        f.write("ENVIRONMENT_VERIFIED")
    print("\n📝 Marker file created: /root/markers/environment_verified.txt")

print("=" * 50)
