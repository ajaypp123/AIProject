
## Setup
```sh
uv init flight-booking-server
# Check flight-booking-server server.py example
cd flight-booking-server
uv add "mcp[cli]"
uv tool install mcp[cli] # install globally
```

## Start server with inspecter
```sh
uv run mcp dev server.py
# Open URL http://localhost:<PORT>/?MCP_PROXY_AUTH_TOKEN=<TOKEN>
```

## Start in vscode
- start server
```json
{
    "flight-booking-mcp-server": {
        "command": "uv",
        "type": "stdio",
        "cwd": "/home/ajayp/code/AIProject/mcp/flight-booking-mcp/flight-booking-server",
        "args": [
            "run",
            "python",
            "server.py"
        ]
	}
}
```

- kill server
```sh
kill -9 $(lsof -t -i:8000)
```
