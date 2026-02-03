# Development Commands

## Build and Development

- `mise build` - Build the application binary (outputs to `./bin/docker-agent`)
- `mise test` - Run Go tests (clears API keys to ensure deterministic tests)
- `mise lint` - Run golangci-lint (uses `.golangci.yml` configuration)
- `mise format` - Format code using golangci-lint fmt
- `mise dev` - Run lint, test, and build in sequence

## Docker and Cross-Platform Builds

- `mise build-local` - Build binary for local platform using Docker Buildx
- `mise cross` - Build binaries for multiple platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64)
- `mise build-image` - Build Docker image tagged as `docker/docker-agent`
- `mise push-image` - Build and push multi-platform Docker image to registry

## Running docker-agent

- `./bin/docker-agent run <config.yaml>` - Run agent with configuration (launches TUI by default)
- `./bin/docker-agent run <config.yaml> -a <agent_name>` - Run specific agent from multi-agent config
- `./bin/docker-agent run agentcatalog/pirate` - Run agent directly from OCI registry
- `./bin/docker-agent run --exec <config.yaml>` - Execute agent without TUI (non-interactive)
- `./bin/docker-agent new` - Generate new agent configuration interactively
- `./bin/docker-agent new --model openai/gpt-5` - Generate with specific model
- `./bin/docker-agent share push ./agent.yaml namespace/repo` - Push agent to OCI registry
- `./bin/docker-agent share pull namespace/repo` - Pull agent from OCI registry
- `./bin/docker agent serve mcp ./agent.yaml` - Expose agents as MCP tools
- `./bin/docker agent serve a2a <config.yaml>` - Start agent as A2A server
- `./bin/docker agent serve api` - Start Docker `docker-agent` API server

## Debug and Development Flags

- `--debug` or `-d` - Enable debug logging (logs to `~/.cagent/cagent.debug.log`)
- `--log-file <path>` - Specify custom debug log location
- `--otel` or `-o` - Enable OpenTelemetry tracing
- Example: `./bin/docker-agent run config.yaml --debug --log-file ./debug.log`

# Testing

- Tests located alongside source files (`*_test.go`)
- Run `mise test` to execute full test suite
- E2E tests in `e2e/` directory
- Test fixtures and data in `testdata/` subdirectories

# Agent's config yaml

- Those config yaml follow a strict schema: ./agent-schema.json
- The schema is versioned.
- ./pkg/config/v0, ./pkg/config/v1... packages handle older versions of the config.
- ./pkg/config/latest packages handles the current, work in progress config format.
- When adding new features to the config, only add them the latest config.
- Older config types are frozen.
- When adding new features to the config, update ./agent-schema.json accordingly and create an example yaml
  that demonstrates the new feature.

This project uses `github.com/stretchr/testify` for assertions.

In Go tests, always prefer `require` and `assert` from the `testify` package over manual error handling.

### Configuration Validation

- All agent references must exist in config
- Model references can be inline (e.g., `openai/gpt-4o`) or defined in models section
- Tool configurations validated at startup

### Adding New Features

- Follow existing patterns in `pkg/` directories
- Implement proper interfaces for providers and tools
- Add configuration support if needed
- Consider both CLI and TUI interface impacts, along with API server impacts

## Model Provider Configuration Examples

Models can be referenced inline (e.g., `openai/gpt-4o`) or defined explicitly:

### OpenAI

```yaml
models:
  gpt4:
    provider: openai
    model: gpt-4o
    temperature: 0.7
    max_tokens: 4000
```

### Anthropic

```yaml
models:
  claude:
    provider: anthropic
    model: claude-sonnet-4-0
    max_tokens: 64000
```

### Gemini

```yaml
models:
  gemini:
    provider: google
    model: gemini-2.0-flash
    temperature: 0.5
```

### DMR

```yaml
models:
  dmr:
    provider: dmr
    model: ai/llama3.2
```

## Tool Configuration Examples

### Local MCP Server (stdio)

```yaml
toolsets:
  - type: mcp
    command: "python"
    args: ["-m", "mcp_server"]
    tools: ["specific_tool"] # optional filtering
    env:
      API_KEY: "value"
```

### Remote MCP Server (SSE)

```yaml
toolsets:
  - type: mcp
    remote:
      url: "http://localhost:8080/mcp"
      transport_type: "sse"
      headers:
        Authorization: "Bearer token"
```

### Docker-based MCP Server

```yaml
toolsets:
  - type: mcp
    ref: docker:github-official
    instruction: |
      Use these tools to help with GitHub tasks.
```

### Memory Tool with Custom Path

```yaml
toolsets:
  - type: memory
    path: "./agent_memory.db"
```

### Shell Tool

```yaml
toolsets:
  - type: shell
```

### Filesystem Tool

```yaml
toolsets:
  - type: filesystem
```

## Common Development Patterns

### Agent Hierarchy Example

```yaml
agents:
  root:
    model: anthropic/claude-sonnet-4-0
    description: "Main coordinator"
    sub_agents: ["researcher", "writer"]
    toolsets:
      - type: transfer_task
      - type: think

  researcher:
    model: openai/gpt-4o
    description: "Research specialist"
    toolsets:
      - type: mcp
        ref: docker:search-tools

  writer:
    model: anthropic/claude-sonnet-4-0
    description: "Writing specialist"
    toolsets:
      - type: filesystem
      - type: memory
        path: ./writer_memory.db
```

### Session Commands During CLI Usage

- `/new` - Clear session history
- `/compact` - Generate summary and compact session history
- `/copy` - Copy the current conversation to the clipboard
- `/eval` - Save evaluation data

## File Locations and Patterns

### Key Package Structure

- `pkg/agent/` - Core agent abstraction and management
- `pkg/runtime/` - Event-driven execution engine
- `pkg/tools/` - Built-in and MCP tool implementations
- `pkg/model/provider/` - AI provider implementations
- `pkg/session/` - Conversation state management
- `pkg/config/` - YAML configuration parsing and validation
- `pkg/gateway/` - MCP gateway/server implementation
- `pkg/tui/` - Terminal User Interface components
- `pkg/api/` - API server implementation

### Configuration File Locations

- `examples/` - Sample agent configurations
- Root directory - Main project configurations (`Taskfile.yml`, `go.mod`)

### Environment Variables

- `OPENAI_API_KEY` - OpenAI authentication
- `ANTHROPIC_API_KEY` - Anthropic authentication
- `GOOGLE_API_KEY` - Google/Gemini authentication
- `MISTRAL_API_KEY` - Mistral authentication
- `TELEMETRY_ENABLED` - Control telemetry (set to false to disable)
- `CAGENT_HIDE_TELEMETRY_BANNER` - Hide telemetry banner message

## Port Allocation

### Port Range Policy

**IMPORTANT**: All new applications and services must use ports in the range **11000-12000**.

- **Reserved Range**: 11000-12000 for all new applications
- **Purpose**: Avoid conflicts with common development ports (3000, 8080, etc.)
- **Allocation**: Choose an available port within this range
- **Documentation**: Document port assignments in project README or configuration

**Examples:**
- Chat Copilot Frontend: 11000
- Chat Copilot Backend: 11001
- Custom API Server: 11002
- Development Tools: 11003+

**Port Checking:**
```powershell
# Check if port is available
Test-NetConnection -ComputerName localhost -Port 11000 -InformationLevel Quiet

# Find available port in range
$port = 11000
while ($port -le 12000) {
    $available = -not (Test-NetConnection -ComputerName localhost -Port $port -InformationLevel Quiet -WarningAction SilentlyContinue)
    if ($available) { break }
    $port++
}
```

## Debugging and Troubleshooting

### Debug Mode

- Add `--debug` flag to any command for detailed logging
- Logs written to `~/.cagent/cagent.debug.log` by default
- Use `--log-file <path>` to specify custom log location
- Example: `./bin/cagent run config.yaml --debug`

### OpenTelemetry Tracing

- Add `--otel` flag to enable OpenTelemetry tracing
- Example: `./bin/cagent run config.yaml --otel`
