from mcp import SamplingMessage
from mcp.server.fastmcp import FastMCP, Context
from mcp.types import TextContent
from pydantic import BaseModel

class UserInfo(BaseModel):
    name: str
    age: int

def register_roots(mcp: FastMCP):
    @mcp.tool()
    async def get_client_roots(ctx: Context):
        """ Assess files from client """
        roots_res = await ctx.session.list_roots()
        if not roots_res:
            return "No workspace roots have been provided by the client application."

        return roots_res

def register_sampling(mcp: FastMCP):
    @mcp.tool()
    async def generate_samples(topic: str, ctx: Context):
        """Generate text via client LLM: request client to fetch details about topic"""
        # Create prompt
        prompt = f"Write something about {topic}"

        print(f"Using prompt for sampling {prompt}")
        # Request result from LLM
        result = await ctx.session.create_message(messages=[
            SamplingMessage(
                role="user",
                content=TextContent(type="text", text=prompt)
        )], max_tokens=300)

        print(f"Result for sampling is {result.content.text}")
        # Return generated result
        return result.content.text

def register_elicitation(mcp: FastMCP):
    @mcp.tool()
    async def ask_user_info(ctx: Context) -> str:
        """ Take input from user for name and age """
        # Requests structured input from user
        result = await ctx.elicit(
            message="Please provide your information",
            schema=UserInfo
        )

        user_info = result.data
        return f"Hello {user_info.name}, {user_info.age}"

def register_tools(mcp: FastMCP):

    @mcp.tool()
    def search_flights(origin: str, destination: str) -> dict:
        """Search for flights between two airports"""
        return {
            "flights": [
                {"id": "FL123", "origin": origin, "destination": destination, "price": 299},
                {"id": "FL456", "origin": origin, "destination": destination, "price": 399}
            ]
        }

    @mcp.tool()
    def create_booking(flight_id: str, passenger_name: str) -> dict:
        """Create a flight booking"""
        return {
            "booking_id": f"BK{flight_id[-3:]}",
            "flight_id": flight_id,
            "passenger": passenger_name,
            "status": "confirmed"
        }
