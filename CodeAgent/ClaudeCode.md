
# Claude

## **Installation and Setup**

```sh
curl -fsSL https://claude.ai/install.sh | bash
```

### Assess with API

- Update details, get from `https://kodekloud.com/ai-playgrounds/kodekey` kodekloud kodekey.

```sh
export ANTHROPIC_MODEL=''
export ANTHROPIC_BASE_URL=''
export ANTHROPIC_AUTH_TOKEN=''
```

- Start claude code
```sh
claude
```

## **Commands**

all commands start with `/`

### help, status, cwd
```sh
/help for help, /status for your current setup

cwd: /Users/jeremy/demos

> <write prompt to create application or work on any task>
> create a python file that reads members.csv and displays first and last names.

```

### `/init`
- Scans the repository for files (e.g., package.json, pyproject.toml, requirements, Python sources).
- Generates a CLAUDE.md file describing the project, usage, and recommended next steps.

### `/terminal-setup`
- Run /terminal-setup to set up terminal integration

### `/model`
- create model application

## **Configuration**

### Global setting
- `~/.claude/settings.json`
```json
{
  "permissions": {
    "allow": [
      "Bash(npm run lint)",
      "Bash(npm run test:*)",
      "Read(~/.zshrc)"
    ],
    "deny": [
      "Bash(curl:*)",
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)"
    ],
    "defaultMode": "acceptEdits",
    "additionalDirectories": ["../some-other-project"]
  },
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": "grpc"
  }
}
```

### Local settings
- `settings.local.json` create at project root level
```
{
  "cleanupAfterCountdown": 7
}
```


## **Using Claude**

### Building a project from scratch

- prompt:
This will create project and ask every file to review and if any permission needed.
```
Create a FLASK web application. This application will be a "METAR Reader". The user can type in an airport code, and then the application will fetch the METAR reading from that airport, and decode it. METAR is a standardized weather report. It is somewhat cryptic so I would like to convert it into plain English that people can understand. For instance, "Clear day, 70 degrees, wind 5mph to the south". This app will be successful if people can type in an airport code, and receive a friendly readable weather report. Use an aviation weather API (for example NOAA/ADDS METAR service, e.g. https://aviationweather.gov/adds/dataserver_current/httpparam?dataSource=metars&requestType=retrieve&stationString=KHIO&format=xml) as a reference.
```

```
Create a complete authentication system for a React application. Include user registration, login, password reset, JWT tokens, email verification, and a simple frontend. Use Express, MongoDB, and modern security best practices.
```

- Development prompt example

```txt
Create a POST /api/auth/login endpoint using Express.js with the following requirements:

Input:
- email (string, required, must be a valid email)
- password (string, required, min 8 characters)

Process:
- Validate input using express-validator
- Check if user exists in PostgreSQL database
- Compare password using bcrypt
- Generate JWT token with 24h expiration

Response:
- Success: { token, user: { id, email, name } }
- Failure: Appropriate error message and status code

Include error handling for database failures and validation errors.
```


### Managing long sessions

- after long coversasation, store SESSION_NOTES.md to keep track.
- It can have multiple session.

```prompt
Help me create SESSION_NOTES.md for this session.
```

- After restart, it will reopen session.

- example:
```md
## Project Overview
Built a Flask app for image optimization:
- Drag-and-drop upload
- JPEG/PNG/WebP support (SVG/GIF rejected)
- Rate limiting 30 req/min per IP
- Max upload 25MB
- Auto-cleanup of temp files
- Before/After UI
- Uses Pillow (OpenCV available)

## SESSION 1: Core features
- Upload, validation, temp file handling
- Rate limiting with Flask-Limiter (in-memory store for dev)
- Error handlers for 413 and 429 responses

## SESSION 2: Advanced features
- Quality slider, resize by percent
- Sharpen/blur controls, contrast/brightness (added)
- Format conversion (JPEG/PNG/WebP)
- Live iterative optimization workflow: upload once, optimize multiple times
- Fix: convert float sharpen values to int percent for UnsharpMask
- UI improvements: responsive layout, presets, live preview

## Development notes
- Use python3 -m venv venv and source venv/bin/activate
- If port 5000 in use on macOS, run on a different port or disable AirPlay Receiver
- For production, configure Redis for rate limiter storage

## Next steps / Roadmap
- Add persistent rate-limiter storage (Redis)
- Add presets persistence and user accounts (requires DB)
- Add server-side caching for repeated identical optimizations
```

## **Code Review**

Ref for prompt:
https://notes.kodekloud.com/docs/Claude-Code-For-Beginners/Code-Review-with-Claude-Code/Demo-Initial-Project-Analysis/page

- Initial Project Analysis
- Design Pattern Implementation Review
- Code Quality Metrics Standards
- Code Duplication Detection
- Naming Conventions Readability
- Test Coverage Quality Analysis

## **Security Auditing**

Write Prompt to improve application security and. This prompt will evaluate and create report for review and can be provide to coding agent.

Ref for prompt:
https://notes.kodekloud.com/docs/Claude-Code-For-Beginners/Security-Auditing-with-Claude-Code/Demo-Initial-Scan/page

- Initial Scan
- Authentication Flow Review
- Database Security Audit
- DemoAPI and Infrastructure Security
- File Handling Auditing
- Business Logic Vulnerabilities
- Comprehensive Report

## **Agents**

Create new agent assign permission and role. It will perform action based on role.

```
> /agents


> Create new agent


> 1. Generate with Claude (recommended)
2. Manual configuration



Select tools
────────────────────────
All tools
Read-only tools
> Edit tools
Execution tools
MCP tools



Select model
  1. Sonnet           Balanced performance -- best for most agents
  2. Haiku            Fast and efficient for simple tasks
> 3. Inherit from parent  Use the same model as the main conversation


Confirm and save

Name: code-readability-reviewer
Location: .claude/agents/code-readability-reviewer.md
Tools: Bash, Glob, Grep, LS, Read, WebFetch, TodoWrite, WebSearch, BashOutput, KillBash
Model: Inherit from parent


> use the code-readability-reviewer agent to inspect auth.js and give me a report.
```

## **MCP**

https://code.claude.com/docs/en/mcp-quickstart
