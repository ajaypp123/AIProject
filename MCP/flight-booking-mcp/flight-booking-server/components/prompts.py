from mcp.server.fastmcp import FastMCP

def register_prompts(mcp: FastMCP):

    @mcp.prompt()
    def find_best_flight(budget: float, preferences: str = "economy") -> str:
        """Generate a prompt for finding the best flight within budget"""
        return f"""Please help me find the best flight within a ${budget} budget.

    My preferences: {preferences}

    Please consider:
    - Price (must be under ${budget})
    - Flight duration
    - Airline reputation
    - Departure times

    Use the search_flights tool to find available options and provide a recommendation with reasoning."""

    @mcp.prompt()
    def handle_disruption(original_flight: str, reason: str) -> str:
        """Generate a prompt for handling flight disruptions"""
        return f"""A passenger's flight {original_flight} has been disrupted due to: {reason}

    Please help resolve this by:
    1. Understanding the passenger's situation
    2. Finding alternative flight options using search_flights
    3. Providing clear rebooking steps
    4. Offering appropriate compensation if applicable

    Be empathetic and solution-focused in your response."""
