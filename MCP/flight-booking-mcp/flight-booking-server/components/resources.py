from mcp.server.fastmcp import FastMCP

def register_resources(mcp: FastMCP):
    @mcp.resource("file://airports")
    def get_airports():
        """Get list of available airports"""
        return {
            "LAX": {"name": "Los Angeles International", "city": "Los Angeles"},
            "JFK": {"name": "John F. Kennedy International", "city": "New York"},
            "LHR": {"name": "London Heathrow", "city": "London"}
        }

    @mcp.resource("file://airlines")
    def get_airlines():
        """Get list of available airlines and their information"""
        return {
            "AA": {"name": "American Airlines", "country": "USA", "fleet_size": 950},
            "BA": {"name": "British Airways", "country": "UK", "fleet_size": 280},
            "DL": {"name": "Delta Air Lines", "country": "USA", "fleet_size": 860},
            "UA": {"name": "United Airlines", "country": "USA", "fleet_size": 790}
        }
