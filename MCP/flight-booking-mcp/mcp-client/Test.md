
## Start mcp server in vscode
```json
	"servers": {
		"flight-booking-mcp-server": {
			"command": "uv",
			"type": "stdio",
			"cwd": "/home/ajayp/code/AIProject/mcp/flight-booking-mcp/flight-booking-server",
			"args": [
				"run",
				"python",
				"server.py",
				"http"
			]
		}
	}
```

## Setup client
```sh
cd /home/lab-user
uv init mcp-client
cd mcp-client
uv add mcp[client]
```

## Run test case
```sh
uv run python <test-case>.py
```
