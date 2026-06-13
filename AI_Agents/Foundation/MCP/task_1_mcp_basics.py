#!/usr/bin/env python3
"""Task 1: Understanding MCP Basics - Your first MCP server"""

import os
import asyncio
from typing import Any

# ╔════════════════════════════════════════╗
# ║     Model Context Protocol (MCP)      ║
# ╚════════════════════════════════════════╝
#
# What is MCP?
# ------------
# MCP is like a USB port for AI - a standard way for
# AI models to connect to external tools and data sources.
#
# ┌─────────────────┐     MCP Protocol    ┌─────────────────┐
# │   AI Assistant  │◄────────────────────►│   MCP Server    │
# │   (LangGraph)   │   stdio/SSE/HTTP    │  (Your Tools)   │
# └─────────────────┘                      └─────────────────┘
#
# MCP Server Components:
# ┌──────────────────────────────────────┐
# │          MCP Server                  │
# ├──────────────────────────────────────┤
# │ 1. Tools (Functions)                 │
# │    └─ add, multiply, divide          │
# │ 2. Resources (Optional)              │
# │    └─ Files, data, configs           │
# │ 3. Prompts (Optional)                │
# │    └─ Pre-defined templates          │
# └──────────────────────────────────────┘

print("📡 Task 1: Understanding MCP Basics\n")

# Import MCP SDK components
try:
    from mcp.server.fastmcp import FastMCP
except ImportError:
    print("⚠️ Creating mock FastMCP for learning (install 'pip install mcp' for real implementation)")

    # Mock implementation for learning purposes
    class FastMCP:
        """Mock FastMCP server for learning"""
        def __init__(self, name):
            self.name = name
            self.tools = []

        def tool(self):
            """Mock tool decorator"""
            def decorator(func):
                self.tools.append({
                    'name': func.__name__,
                    'function': func
                })
                return func
            return decorator

        def run(self, transport="stdio"):
            print(f"🚀 {self.name} MCP Server would run with {transport} transport")
            print(f"📦 Available tools: {[t['name'] for t in self.tools]}")

# TODO 1: Initialize the MCP server
# Hint: Pass server name to FastMCP with double quotes
mcp = FastMCP("___")  # Replace ___ with Calculator

# Create calculator tools using FastMCP decorators
@mcp.tool()
def add(a: float, b: float) -> float:
    """Add two numbers together"""
    result = a + b
    print(f"  🔧 Tool 'add' called with a={a}, b={b}")
    print(f"  ➕ Result: {result}")
    return result

# TODO 2: Create the multiply tool
# Hint: Use @mcp.tool() decorator
___  # Replace ___ with @mcp.tool()
def multiply(a: float, b: float) -> float:
    """Multiply two numbers"""
    result = a * b
    print(f"  🔧 Tool 'multiply' called with a={a}, b={b}")
    print(f"  ✖️ Result: {result}")
    return result

@mcp.tool()
def divide(a: float, b: float) -> str:
    """Divide two numbers with zero check"""
    print(f"  🔧 Tool 'divide' called with a={a}, b={b}")

    if b == 0:
        print("  ❌ Error: Division by zero!")
        return "Error: Cannot divide by zero"

    result = a / b
    print(f"  ➗ Result: {result}")
    return f"{a} ÷ {b} = {result}"

# Test the tools directly (simulating MCP calls)
print("\n" + "=" * 60)
print("TESTING MCP TOOLS:")
print("=" * 60)

def test_tools():
    """Test our MCP tools directly"""

    # Test addition
    print("\nTest 1: Addition")
    result = add(5, 3)
    print(f"Response: {result}")

    # Test multiplication
    print("\nTest 2: Multiplication")
    result = multiply(4, 7)
    print(f"Response: {result}")

    # Test division
    print("\nTest 3: Division")
    result = divide(10, 2)
    print(f"Response: {result}")

    # Test division by zero
    print("\nTest 4: Division by zero")
    result = divide(5, 0)
    print(f"Response: {result}")

# Run the tests
test_tools()

if __name__ == "__main__":
    # Create marker file (must happen before server starts)
    os.makedirs("/root/markers", exist_ok=True)
    with open("/root/markers/task1_mcp_basics_complete.txt", "w") as f:
        f.write("TASK1_COMPLETE")

    print("\n" + "=" * 60)
    print("💡 KEY CONCEPTS:")
    print("- FastMCP creates MCP servers easily")
    print("- @mcp.tool() decorator exposes functions")
    print("- Tools have type hints for parameters")
    print("- Servers can run via stdio, SSE, or HTTP")
    print("- AI models call tools through MCP protocol")
    print("=" * 60)

    print("\n✅ Task 1 complete! MCP tools tested successfully.")
    print("\n" + "=" * 60)
    print("🚀 STARTING MCP SERVER")
    print("=" * 60)
    print("The calculator MCP server is now starting...")
    print("Keep this terminal open - the server will run continuously.")
    print("Use Ctrl+C to stop the server when you're done.")
    print("\nServer ready! Waiting for client connections...")
    print("=" * 60 + "\n")

    # Start the MCP server (blocks indefinitely)
    # This is the production pattern - servers run continuously
    mcp.run(transport="stdio")