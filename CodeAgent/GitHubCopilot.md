
# Introduction

## Pair Programming

- Github Copilot is agent which can do pair programming.
- Pattern Recognition, Instant suggestion, Documentation.

Use case:

- Boilerplate and repetitive task.
- API Integration Pattern.
- Test Generation
- Document Writing.
- Code Assistant.

## Strength and Limitation

Strength

- Increased Productivity
- Reduced Load
- Faster prototyping
- Documentation build
- Learning tool

Limitation

- Code Quality
- Vulnerability
- Token Limit
- Performance

## SDLC Life cycle

1. Planning
2. Analysis
3. Design
4. Implementation
5. Testing and Integration
6. Maintenance

## Instruction for copilot
- Provide instruction followed by copilot like coding standreds, ref, instructions.
```
mkdir -p .github/copilot
touch .github/copilot/instruction.md
```

## copilot Chat command

1. /new
- Create new application
eg: /new Create new python app to store data in memory db.

2. /fix
- fix selected context
eg: /fix fix this code app not working.

3. /doc
Insert documentation in code

4. /exp
Start new context with fresh thread

5. /explain
Provide explanation of thread.

6. fixTestFailure
Fix unit test

7. /generate
Generate snipit based on question.

8. /help
provide help.

9. /optimize
optimize code

10. /tests

## Github Agent

- Agent can be invoke by @

@workspace
@terminal
@vscode
@azure
@github
