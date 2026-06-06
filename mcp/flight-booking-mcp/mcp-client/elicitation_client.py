#!/usr/bin/env python3
"""
Elicitation MCP Client - Handle server user input requests
"""
import asyncio
from mcp.client.session import ClientSession
from mcp.client.streamable_http import streamablehttp_client
from mcp.types import ElicitRequestParams, ElicitResult

async def handle_elicitation(context, params: ElicitRequestParams) -> ElicitResult:
    """Handle elicitation requests from server - request actual user input"""
    print(f"🔔 Server asks for user input: {params.message}")
    print("=" * 50)

    # Request actual user input
    try:
        user_response = input("👤 Please enter your response: ").strip()

        if not user_response:
            print("⚠️ No input provided, declining request")
            return ElicitResult(action="decline", content={"reason": "No input provided"})

        # Parse user response - try to handle common formats
        response_data = {}

        # Check if user provided JSON-like response
        if user_response.startswith('{') and user_response.endswith('}'):
            try:
                import json
                response_data = json.loads(user_response)
                print(f"✅ Parsed JSON response: {response_data}")
            except json.JSONDecodeError:
                response_data = {"response": user_response}
                print(f"📝 Using raw response: {user_response}")
        else:
            # Handle simple text responses intelligently based on the request
            message_lower = params.message.lower()

            if "name" in message_lower:
                response_data = {"name": user_response}
            elif "email" in message_lower:
                response_data = {"email": user_response}
            elif "budget" in message_lower or "price" in message_lower:
                try:
                    budget_value = float(user_response.replace('$', '').replace(',', ''))
                    response_data = {"budget": budget_value, "currency": "USD"}
                except ValueError:
                    response_data = {"budget_text": user_response}
            elif "confirm" in message_lower or "approve" in message_lower:
                is_confirmed = user_response.lower() in ['yes', 'y', 'true', '1', 'confirm', 'accept']
                response_data = {"confirmation": "yes" if is_confirmed else "no"}
            else:
                response_data = {"response": user_response}

        print(f"✅ User input accepted: {response_data}")
        return ElicitResult(action="accept", content=response_data)

    except KeyboardInterrupt:
        print("\n❌ User cancelled input")
        return ElicitResult(action="decline", content={"reason": "User cancelled"})
    except Exception as e:
        print(f"❌ Error processing user input: {e}")
        return ElicitResult(action="decline", content={"reason": f"Input error: {str(e)}"})

async def test_elicitation():
    """Test MCP elicitation - server requests user input through client"""
    print("🔔 ELICITATION MCP CLIENT")
    print("=" * 40)
    print("🎯 Goal: Handle server user input requests")
    print("=" * 40)

    try:
        async with streamablehttp_client("http://localhost:8000/mcp/") as (read, write, _):
            # Connect with elicitation support
            async with ClientSession(read, write, elicitation_callback=handle_elicitation) as client:
                await client.initialize()
                print("✅ Connected to server with elicitation support!")
                print()

                # List available tools to see if server has user-interactive tools
                print("🔧 Checking for interactive tools...")
                tools = await client.list_tools()

                interactive_tools = []
                for tool in tools.tools:
                    if any(keyword in tool.description.lower() for keyword in ['user', 'input', 'ask', 'collect', 'request']):
                        interactive_tools.append(tool)
                        print(f"  🤝 {tool.name}: {tool.description}")

                if not interactive_tools:
                    print("  ⚠️  No obvious interactive tools found")
                    print("     This is normal - the basic flight server doesn't have user input tools")
                print()

                # Test standard tools to see if they trigger elicitation
                print("🧪 Testing for potential elicitation triggers...")

                try:
                    print("🎫 Testing booking creation (might ask for user details)...")
                    booking_result = await client.call_tool("create_booking", {
                        "flight_id": "FL456",
                        "passenger_name": "Test User"
                    })
                    if booking_result.content and len(booking_result.content) > 0:
                        print(f"✅ Booking result: {booking_result.content[0].text}")
                    print()
                except Exception as e:
                    print(f"⚠️  Booking test: {e}")
                    print()

                # Demonstrate elicitation capability
                print("🔔 Elicitation callback is ready!")
                print("   If the server had tools that use ctx.session.elicit(),")
                print("   our elicitation callback would handle those requests.")
                print()
                print("📋 Elicitation callback features:")
                print("   - Prompts user for real input when server requests")
                print("   - Handles text responses and JSON input")
                print("   - Intelligently parses responses based on request type")
                print("   - Allows user to decline with Ctrl+C")
                print("   - Supports various input formats")
                print()

                print("🎯 Example user interaction:")
                print("   • Server asks: 'Please enter your name'")
                print("   • Client prompts: 'Please enter your response: '")
                print("   • User types: 'John Smith'")
                print("   • Client sends: {'name': 'John Smith'}")
                print()
                print("💡 Supported input formats:")
                print("   • Simple text: 'John Smith' → {'name': 'John Smith'}")
                print("   • JSON: '{\"name\": \"John\", \"age\": 30}' → parsed as JSON")
                print("   • Numbers: '500' for budget → {'budget': 500.0, 'currency': 'USD'}")
                print("   • Confirmations: 'yes' → {'confirmation': 'yes'}")
                print()

                print("🎉 Elicitation client test completed!")
                print("✨ Ready to handle server user input requests!")

    except Exception as e:
        print(f"❌ Elicitation client failed: {e}")
        print("💡 Make sure the flight booking server is running on port 8000")

if __name__ == "__main__":
    asyncio.run(test_elicitation())