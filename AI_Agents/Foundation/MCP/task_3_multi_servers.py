#!/usr/bin/env python3
"""Task 3: Multiple MCP Servers - Orchestrating calculator and weather servers"""

import os
import asyncio
from langchain.agents import create_agent
from langchain_openai import ChatOpenAI

# ╔════════════════════════════════════════════╗
# ║     Multiple MCP Servers Architecture      ║
# ╚════════════════════════════════════════════╝
#
#              [User Query]
#                    │
#                    ▼
#            ┌──────────────┐
#            │  LangGraph   │
#            │ React Agent  │
#            └──────┬───────┘
#                   │
#         ┌─────────┴─────────┐
#         │MultiServerMCPClient│
#         └─────────┬─────────┘
#                   │
#      ┌────────────┴────────────┐
#      ▼                         ▼
# ┌──────────┐            ┌──────────┐
# │Calculator│            │ Weather  │
# │MCP Server│            │MCP Server│
# │    🔢    │            │    ☁️    │
# └──────────┘            └──────────┘
#
# The agent automatically routes to the appropriate server
# based on the query content and available tools

print("🌐 Task 3: Multiple MCP Servers\n")

# Import MCP adapter - uses real langchain-mcp-adapters package
from langchain_mcp_adapters.client import MultiServerMCPClient

# Initialize the LLM
model = ChatOpenAI(
    model=os.getenv("OPENAI_MODEL", "openai/gpt-4.1-mini"),
    base_url=os.getenv("OPENAI_API_BASE"),
    api_key=os.getenv("OPENAI_API_KEY"),
    temperature=0
)

print("Configuring multiple MCP servers:\n")

# TODO 1: Initialize MultiServerMCPClient with both servers
# Hint: Add both calculator and weather server configurations
client = MultiServerMCPClient(
    {
        ___: {  # Replace ___ with "calculator"
            "command": "python",
            "args": ["/root/code/mcp_servers/calculator_server.py"],
            "transport": "stdio",
        },
        "weather": {
            "command": "python",
            "args": ["/root/code/mcp_servers/weather_server.py"],
            "transport": ___,  # Replace ___ with "stdio"
        }
    }
)

async def run_multi_server_agent():
    """Create and run agent with multiple MCP servers"""

    print("📦 Loading tools from multiple servers...")

    # TODO 2: Get all tools from both servers
    # Hint: Use client.get_tools()
    tools = await ___  # Replace ___ with client.get_tools()

    print(f"✅ Loaded {len(tools) if hasattr(tools, '__len__') else 'multiple'} tools from MCP servers")

    # TODO 3: Create react agent with all tools
    # Hint: Pass model and tools to create_agent
    agent = create_agent(___, ___)  # Replace both ___ with model, tools

    print("\n" + "=" * 60)
    print("TESTING MULTI-SERVER ORCHESTRATION:")
    print("=" * 60)

    # Test 1: Calculator query
    print("\nTest 1: Calculator MCP")
    calc_response = await agent.ainvoke({
        "messages": "What is 42 plus 58?"
    })
    print(f"Response: {calc_response['messages'][-1].content}")

    # Test 2: Weather query
    print("\nTest 2: Weather MCP")
    weather_response = await agent.ainvoke({
        "messages": "What's the weather in London?"
    })
    print(f"Response: {weather_response['messages'][-1].content}")

    # Test 3: Complex math
    print("\nTest 3: Complex Math")
    complex_response = await agent.ainvoke({
        "messages": "What's (3 + 5) x 12?"
    })
    print(f"Response: {complex_response['messages'][-1].content}")

    # Test 4: Another weather query
    print("\nTest 4: Weather in Multiple Cities")
    cities_response = await agent.ainvoke({
        "messages": "Compare the weather in New York and Tokyo"
    })
    print(f"Response: {cities_response['messages'][-1].content}")

    # Test 5: Mixed query
    print("\nTest 5: Mixed Query")
    mixed_response = await agent.ainvoke({
        "messages": "If it's 20°C in Paris and temperature rises by 5 degrees, what will it be?"
    })
    print(f"Response: {mixed_response['messages'][-1].content}")

# Run the multi-server agent
if __name__ == "__main__":
    print("Starting Multi-Server MCP Orchestration...")
    print("This demonstrates how a single agent can use multiple MCP servers\n")

    # Run async function
    asyncio.run(run_multi_server_agent())

    print("\n" + "=" * 60)
    print("💡 KEY CONCEPTS:")
    print("- MultiServerMCPClient manages multiple MCP servers")
    print("- Each server exposes different tools")
    print("- Agent automatically selects appropriate tools")
    print("- Seamless orchestration across servers")
    print("- Extensible to many servers and domains")
    print("=" * 60)

    # Create marker file
    os.makedirs("/root/markers", exist_ok=True)
    with open("/root/markers/task3_multi_servers_complete.txt", "w") as f:
        f.write("TASK3_COMPLETE")