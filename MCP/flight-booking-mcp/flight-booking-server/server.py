import sys
from mcp.server.fastmcp import FastMCP

from components.tools import *
from components.resources import register_resources
from components.prompts import register_prompts

# Create an MCP server
mcp = FastMCP("Flight Booking Server")

# Server components
register_tools(mcp)
register_resources(mcp)
register_prompts(mcp)

# Client components
register_roots(mcp)
register_sampling(mcp)
register_elicitation(mcp)

if __name__ == "__main__":
    if sys.argv[1] == "http":
        # Run in stdio mode for testing with basic client
        mcp.run(transport="streamable-http")
    elif sys.argv[1] == "sse":
        # Run in streamable HTTP mode for client connections
        mcp.run(transport="sse")
    else:
        mcp.run(transport="sse")
