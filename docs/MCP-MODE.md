# Running Docker `cagent` in MCP Mode

## Why use `cagent mcp`?

The `cagent mcp` command allows your agents to be consumed by other MCP-compatible products and tools. This enables seamless integration with existing workflows and applications that support the Model Context Protocol (MCP).

**Important:** MCP is not just about tools - it's also about agents. By exposing your Docker `cagent` configurations through MCP, you make your specialized agents available to any MCP client, whether that's VS Code, GitHub Copilot CLI, Claude Desktop, Claude Code, or any other MCP-compatible application.

This means you can:
- Use your custom agents directly within VS Code, GitHub Copilot CLI, Claude Desktop, or Claude Code
- Share agents across different applications
- Build reusable agent teams that can be consumed anywhere MCP is supported
- Integrate domain-specific agents into your existing development workflows

## How `cagent mcp` maps agents to MCP tools

When you run `cagent mcp`, each exposed agent becomes an MCP tool. That means:

- A single-agent config becomes a single MCP tool.
- A multi-agent config exposes one tool per agent.
- Use `--agent <name>` when you want to expose only one agent instead of the full team.

Examples:

```bash
# Expose every agent in a local config over stdio
cagent mcp ./examples/dev-team.yaml

# Expose a single agent from that config
cagent mcp ./examples/dev-team.yaml --agent engineer

# Expose an OCI-published agent
cagent mcp agentcatalog/pirate
```

Most desktop MCP clients work best with the default stdio transport. Use `--http` and `--port` only when your MCP client expects a network endpoint instead of launching a local process.

## Common AI client config files in home-directory dot folders

When you want `cagent` available across multiple local AI clients, the most common setup is to point each client at the same `cagent mcp` command and agent file.

The following locations and config roots were either verified locally in this environment or confirmed in upstream documentation:

| Client | Config file | Root key / shape | Status |
| --- | --- | --- | --- |
| GitHub Copilot CLI | `~/.copilot/mcp-config.json` | `mcpServers` object | Verified locally |
| VS Code | `~/.vscode/mcp.json` or workspace `.vscode/mcp.json` | `servers` object | Verified locally and documented by VS Code |
| VS Code Insiders | `~/.vscode-insiders/mcp.json` | `servers` object | Verified locally, follows VS Code format |
| Cursor | `~/.cursor/mcp.json` | `mcpServers` object | Verified locally |
| Cline | `~/.cline/mcp.json` | `servers` object | Verified locally |
| Roo | `~/.roo/mcp.json` | `mcpServers` object | Verified locally |
| Continue | `~/.continue/config.yaml` | `mcpServers` YAML list | Verified locally |
| Gemini CLI | `~/.gemini/settings.json` | `mcpServers` object | Verified locally and documented by Gemini CLI |
| Windsurf | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` object | Verified locally and documented by Windsurf |
| Claude Desktop | `%APPDATA%\\Claude\\claude_desktop_config.json` | `mcpServers` object | Verified locally and documented by Anthropic |
| Antigravity | `~/.antigravity/mcp_config.json` | `mcpServers` object | Locally inferred, not officially documented |

### Recommended shared Windows command

If you already have a local `cagent.exe`, a reusable Windows entry looks like this:

```json
{
  "command": "C:\\Users\\you\\cagent\\bin\\cagent.exe",
  "args": [
    "mcp",
    "C:\\Users\\you\\cagent\\golang_developer.yaml",
    "--working-dir",
    "C:\\Users\\you\\cagent"
  ]
}
```

You can adapt that same command block to each client's config format.

## Using Docker `cagent` agents in VS Code

VS Code MCP servers are configured in `mcp.json`:

- Workspace: `.vscode/mcp.json`
- User profile: use the `MCP: Open User Configuration` command

Add a stdio MCP server entry that launches `cagent mcp`:

```json
{
  "servers": {
    "cagent-dev-team": {
      "type": "stdio",
      "command": "cagent",
      "args": [
        "mcp",
        "agentcatalog/pirate"
      ],
      "env": {
        "OPENAI_API_KEY": "${env:OPENAI_API_KEY}",
        "ANTHROPIC_API_KEY": "${env:ANTHROPIC_API_KEY}"
      }
    }
  }
}
```

To expose a local team config from the current repository, replace `agentcatalog/pirate` with your config path and add `--working-dir`:

```json
{
  "servers": {
    "cagent-dev-team": {
      "type": "stdio",
      "command": "cagent",
      "args": [
        "mcp",
        "C:\\Users\\you\\src\\cagent\\examples\\dev-team.yaml",
        "--working-dir",
        "C:\\Users\\you\\src\\cagent"
      ]
    }
  }
}
```

Notes:

- If `cagent` is not on your `PATH`, set `command` to the full path to `cagent` or `cagent.exe`.
- Keep secrets out of source control. Prefer environment variables or VS Code input variables over hard-coded API keys.
- If your workspace config or prompts contain Windows host paths but `cagent` runs inside WSL or a container, pass `CAGENT_PATH_MAP` so document tools and RAG paths resolve to the mounted runtime paths.
- After saving `mcp.json`, trust and enable the server in VS Code. The exposed agents then show up as tools in chat.

Example:

```json
{
  "servers": {
    "cagent-doc-planner": {
      "type": "stdio",
      "command": "cagent",
      "args": [
        "mcp",
        "C:\\Users\\you\\src\\cagent\\examples\\document_analysis.yaml",
        "--working-dir",
        "C:\\Users\\you\\src\\cagent"
      ],
      "env": {
        "OPENAI_API_KEY": "${env:OPENAI_API_KEY}",
        "CAGENT_PATH_MAP": "C:\\Users\\you\\src\\cagent=/workspace/cagent"
      }
    }
  }
}
```

## Using Docker `cagent` agents in GitHub Copilot CLI

GitHub Copilot CLI can manage custom MCP servers directly from the terminal UI. The simplest setup is:

1. Start `copilot` in the repository where you want to use the agents.
2. Run `/mcp`.
3. Add a new MCP server that uses the `stdio` transport.
4. Set the command to `cagent` (or the full path to `cagent.exe` on Windows if it is not on your `PATH`).
5. Set the arguments to the same `cagent mcp ...` values you would use in a terminal.

Typical argument lists:

```text
mcp agentcatalog/pirate
```

```text
mcp C:\Users\you\src\cagent\examples\dev-team.yaml --working-dir C:\Users\you\src\cagent
```

Pass through whichever provider secrets your agent needs, such as `OPENAI_API_KEY` or `ANTHROPIC_API_KEY`. Once saved, the exposed `cagent` agents become available to Copilot CLI as MCP tools.

## Using Docker `cagent` agents in Gemini CLI

Gemini CLI uses `~/.gemini/settings.json` for user-wide settings and `.gemini/settings.json` for project-specific overrides.

Gemini CLI reads MCP servers from the top-level `mcpServers` object. A typical local `cagent` entry looks like this:

```json
{
  "mcpServers": {
    "cagent": {
      "type": "stdio",
      "command": "C:\\Users\\you\\cagent\\bin\\cagent.exe",
      "args": [
        "mcp",
        "C:\\Users\\you\\cagent\\golang_developer.yaml",
        "--working-dir",
        "C:\\Users\\you\\cagent"
      ],
      "cwd": "C:\\Users\\you\\cagent",
      "env": {
        "ANTHROPIC_API_KEY": "$ANTHROPIC_API_KEY",
        "OPENAI_API_KEY": "$OPENAI_API_KEY"
      }
    }
  }
}
```

Notes:

- Gemini CLI sanitizes sensitive environment variables before spawning MCP child processes.
- If your `cagent` agent needs provider credentials, explicitly pass them in the MCP server's `env` block rather than assuming they will be inherited.
- Gemini CLI also supports remote MCP servers through `url` and `httpUrl`, but local `cagent` integrations are usually simplest with stdio.

## Using Docker `cagent` agents in Windsurf

Windsurf Cascade uses `~/.codeium/windsurf/mcp_config.json` with a top-level `mcpServers` object.

```json
{
  "mcpServers": {
    "cagent": {
      "transport": "stdio",
      "command": "C:\\Users\\you\\cagent\\bin\\cagent.exe",
      "args": [
        "mcp",
        "C:\\Users\\you\\cagent\\golang_developer.yaml",
        "--working-dir",
        "C:\\Users\\you\\cagent"
      ],
      "disabled": false,
      "env": {}
    }
  }
}
```

Notes:

- Windsurf supports `stdio`, `SSE`, and streamable HTTP MCP transports.
- The Windsurf config also supports `${env:VAR_NAME}` and `${file:/path/to/file}` interpolation in `command`, `args`, `env`, `serverUrl`, `url`, and `headers`.
- Team and enterprise deployments can additionally whitelist or centrally manage allowed MCP servers.

## Using Docker `cagent` agents in Cursor, Roo, Cline, and Continue

Several AI clients follow a broadly similar local MCP pattern, but use different file names and root keys:

- Cursor: `~/.cursor/mcp.json` with `mcpServers`
- Roo: `~/.roo/mcp.json` with `mcpServers`
- Cline: `~/.cline/mcp.json` with `servers`
- Continue: `~/.continue/config.yaml` with an `mcpServers` YAML list

Examples:

```json
{
  "mcpServers": {
    "cagent": {
      "type": "stdio",
      "command": "C:\\Users\\you\\cagent\\bin\\cagent.exe",
      "args": [
        "mcp",
        "C:\\Users\\you\\cagent\\golang_developer.yaml",
        "--working-dir",
        "C:\\Users\\you\\cagent"
      ],
      "env": {}
    }
  }
}
```

```json
{
  "servers": {
    "cagent": {
      "type": "stdio",
      "command": "C:\\Users\\you\\cagent\\bin\\cagent.exe",
      "args": [
        "mcp",
        "C:\\Users\\you\\cagent\\golang_developer.yaml",
        "--working-dir",
        "C:\\Users\\you\\cagent"
      ],
      "env": {}
    }
  }
}
```

```yaml
mcpServers:
  - name: cagent
    command: C:\Users\you\cagent\bin\cagent.exe
    args:
      - mcp
      - C:\Users\you\cagent\golang_developer.yaml
      - --working-dir
      - C:\Users\you\cagent
```

## Using Docker `cagent` agents in Antigravity

In this environment, `~/.antigravity/mcp_config.json` appears to be a valid place to store MCP server definitions using a Windsurf-like `mcpServers` object.

However, unlike Gemini CLI and Windsurf, this location is **locally inferred** rather than confirmed by public product documentation.

A cautious configuration looks like this:

```json
{
  "mcpServers": {
    "cagent": {
      "transport": "stdio",
      "command": "C:\\Users\\you\\cagent\\bin\\cagent.exe",
      "args": [
        "mcp",
        "C:\\Users\\you\\cagent\\golang_developer.yaml",
        "--working-dir",
        "C:\\Users\\you\\cagent"
      ],
      "disabled": false,
      "env": {}
    }
  }
}
```

Treat this as an experimental local integration until stronger Antigravity documentation becomes available.

## Using Docker `cagent` agents in Claude Desktop

To use your Docker `cagent` agents in Claude Desktop, add a configuration to your Claude Desktop MCP settings file:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

Here's an example configuration:

```json
{
  "mcpServers": {
    "myagent": {
      "command": "/Users/dockereng/bin/cagent",
      "args": ["mcp", "dockereng/myagent", "--working-dir", "/Users/dockereng/src"],
      "env": {
        "PATH": "/Applications/Docker.app/Contents/Resources/bin:${PATH}",
        "ANTHROPIC_API_KEY": "your_anthropic_key_here",
        "OPENAI_API_KEY": "your_openai_key_here"
      }
    }
  }
}
```

### Configuration breakdown:

- **command**: Full path to your `cagent` binary
- **args**: The MCP command arguments:
  - `mcp`: The subcommand to run Docker `cagent` in MCP mode
  - `dockereng/myagent`: Your agent configuration (can be a local file path or OCI reference)
  - `--working-dir`: Optional working directory for the agent
- **env**: Environment variables needed by your agents:
  - `PATH`: Include any additional paths needed (e.g., Docker binaries)
  - `ANTHROPIC_API_KEY`: Required if your agents use Anthropic models
  - `OPENAI_API_KEY`: Required if your agents use OpenAI models
  - Add any other API keys your agents need (GOOGLE_API_KEY, XAI_API_KEY, etc.)

After updating the configuration, restart Claude Desktop. Your agents will now appear as available tools in Claude Desktop's interface.

## Using Docker `cagent` agents in Claude Code

To add your Docker `cagent` agents to Claude Code, use the `claude mcp add` command:

```bash
claude mcp add --transport stdio myagent \
  --env OPENAI_API_KEY=$OPENAI_API_KEY \
  --env ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  -- cagent mcp agentcatalog/pirate --working-dir $(pwd)
```

### Command breakdown:

- `claude mcp add`: Claude Code command to add an MCP server
- `--transport stdio`: Use stdio transport (standard for local MCP servers)
- `myagent`: Name for this MCP server in Claude Code
- `--env`: Pass through required environment variables (repeat for each variable)
- `--`: Separates Claude Code arguments from the MCP server command
- `cagent mcp agentcatalog/pirate`: The Docker `cagent` MCP command with your agent reference
- `--working-dir $(pwd)`: Set the working directory for the agent

After adding the MCP server, your agents will be available as tools within Claude Code sessions.

## Agent references

You can specify your agent configuration in several ways:

```bash
# Local file path
cagent mcp ./examples/dev-team.yaml

# OCI artifact from Docker Hub
cagent mcp agentcatalog/pirate

# OCI artifact with namespace
cagent mcp dockereng/myagent
```

## Additional options

The `cagent mcp` command supports additional options:

- `--agent <name>`: Expose a single agent instead of all agents in the config
- `--working-dir <path>`: Set the working directory for agent execution
- `--env-from-file <path>`: Load environment variables from one or more files
- `--http`: Serve MCP over streaming HTTP instead of stdio
- `--port <port>`: Choose the HTTP port when using `--http`

Global flags that are especially useful when debugging MCP startup:

- `--debug`: Enable debug logging
- `--log-file <path>`: Write debug logs to a specific file

## Example: Multi-agent team in MCP

When you expose a multi-agent team configuration via MCP, each agent becomes a separate tool. For example, with this configuration:

```yaml
agents:
  root:
    model: claude-sonnet-4-0
    description: "Main coordinator agent"
    instruction: "You coordinate tasks and delegate to specialists"
    sub_agents: ["designer", "engineer"]

  designer:
    model: gpt-5-mini
    description: "UI/UX design specialist"
    instruction: "You create user interface designs and mockups"

  engineer:
    model: claude-sonnet-4-0
    description: "Software engineering specialist"
    instruction: "You implement code based on requirements"
```

All three agents (`root`, `designer`, and `engineer`) will be available as separate tools in the MCP client, allowing you to interact with specific specialists directly or use the root coordinator.

## Troubleshooting

### Agents not appearing in the MCP client

1. Verify your `cagent` binary path is correct
2. Check that all required API keys are set in the environment variables
3. Make sure the MCP server is trusted and enabled in the client
4. Restart or reload the MCP client after configuration changes
5. Check the client logs for connection errors
6. If needed, start `cagent` with `--debug --log-file <path>` and inspect the log output
7. If the client uses a dot-folder config file, confirm you updated the right file and the right root key (`servers` vs `mcpServers`)

### Permission errors

Make sure your `cagent` binary has execute permissions:

```bash
chmod +x /path/to/cagent
```

On Windows, make sure you are pointing at a valid `cagent.exe` path when using an absolute `command`.

### Working directory issues

If your agents need access to specific files or directories, ensure the `--working-dir` parameter points to the correct location and that the agent has appropriate permissions.
