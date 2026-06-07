#!/usr/bin/env python3
"""
Sampling MCP Client - Handle server LLM requests
"""
import asyncio
from mcp.client.session import ClientSession
from mcp.client.streamable_http import streamablehttp_client
from mcp.types import CreateMessageRequestParams, CreateMessageResult, TextContent

async def llm_handler(context, params):
    """
    Handles sampling/createMessage requests from the server.
    """

    print("\n=== Sampling Request Received ===")

    prompt = ""

    for msg in params.messages:
        if hasattr(msg.content, "text"):
            prompt += msg.content.text

    print(f"Prompt: {prompt}")

    generated_text = """
Flight Booking Application

A flight booking application enables users to search, compare,
and reserve airline tickets.

Key Features:
- Search flights by source, destination, and travel date
- Compare fares across airlines
- Book one-way or round-trip journeys
- Manage bookings and cancellations
- Secure online payment integration
- Real-time flight status updates

Benefits:
- Convenient ticket booking
- Faster reservation process
- Centralized booking management
- Better customer experience
"""

    print("Returning mock LLM response...")

    return CreateMessageResult(
        role="assistant",
        content=TextContent(
            type="text",
            text=generated_text
        ),
        model="mock-flight-model"
    )

async def test_sampling():
    """Test MCP sampling - server requests LLM generation through client"""
    print("🎭 SAMPLING MCP CLIENT")
    print("=" * 40)
    print("🎯 Goal: Handle server LLM sampling requests")
    print("=" * 40)

    try:
        async with streamablehttp_client("http://localhost:8000/mcp/") as (read, write, _):
            # Connect with sampling support
            async with ClientSession(read, write, sampling_callback=llm_handler) as client:
                await client.initialize()
                print("✅ Connected to server with sampling support!")
                print()

                # List available tools to see if server has sampling-enabled tools
                print("🔧 Checking for sampling-enabled tools...")
                tools = await client.list_tools()

                sampling_tools = []
                for tool in tools.tools:
                    if any(keyword in tool.description.lower() for keyword in ['generate', 'create', 'write', 'compose']):
                        sampling_tools.append(tool)
                        print(f"  📝 {tool.name}: {tool.description}")

                if not sampling_tools:
                    print("  ⚠️  No obvious sampling-enabled tools found")
                    print("     This is normal - the basic flight server doesn't have LLM tools")
                print()

                # Try to trigger sampling by calling tools that might use LLM
                print("🧪 Testing potential sampling scenarios...")

                # Test existing prompts (they might trigger sampling internally)
                try:
                    print("💡 Testing flight recommendation prompt...")
                    prompt_result = await client.call_tool("generate_samples",
                        { "topic": "Flight Booking System" })

                    print(f"prompt result: {prompt_result}")
                    print("✅ Prompt generated successfully (no sampling required)")
                    print()
                except Exception as e:
                    print(f"⚠️  Prompt test: {e}")
                    print()

                # Demonstrate sampling capability

                print("📋 Sampling callback features:")
                print("   - Handles travel explanation requests")
                print("   - Provides travel/flight explanations")
                print("   - Creates stories and recommendations")
                print("   - Returns contextual responses")
                print()

                print("🎉 Sampling client test completed!")
                print("✨ Ready to handle server LLM requests!")

    except Exception as e:
        print(f"❌ Sampling client failed: {e}")
        print("💡 Make sure the flight booking server is running on port 8000")

if __name__ == "__main__":
    asyncio.run(test_sampling())
