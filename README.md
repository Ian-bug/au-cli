# au

> **Alpha software.** Expect bugs, broken provider endpoints, and rough edges. Use at your own risk.

A minimal AI coding agent for the terminal, built to run on hardware that nothing else will.

## Why

Every AI coding tool I tried — Cursor, Copilot, Claude Desktop, Aider, Continue — is an Electron app, a VS Code extension, or a Python package with 300 transitive dependencies. None of them run comfortably on a 1 GB VPS. I needed something that fits in a single static binary, starts in milliseconds, uses maybe 8 MB of RAM at idle, and still gives me a full agentic coding loop with filesystem access. So I built `au`. It connects to any OpenAI-compatible API endpoint, which means you can point it at a cheap hosted model and get a capable agent running on the smallest DigitalOcean or Hetzner box.

## Install (Linux)

One-liner — downloads the binary, makes it executable, moves it to your PATH:

```sh
curl -fsSL https://github.com/cfpy67/au-cli/releases/latest/download/au-linux-amd64 -o au && chmod +x au && sudo mv au /usr/local/bin/au
```

For ARM64 (Raspberry Pi, Ampere VPS):

```sh
curl -fsSL https://github.com/cfpy67/au-cli/releases/latest/download/au-linux-arm64 -o au && chmod +x au && sudo mv au /usr/local/bin/au
```

Then just run:

```sh
au
```

## Build from source

Requires Go 1.22+:

```sh
git clone https://github.com/cfpy67/au-cli
cd au-cli
go build -o au .
```

## Usage

On first run, use `/connect` to pick a provider and model.

## Commands

| Command | Description |
|---|---|
| `/connect` | Interactive provider + model setup wizard |
| `/use <n>` | Switch to provider number `n` from the list |
| `/key <k>` | Set API key for current provider |
| `/model <m>` | Switch to model `m` |
| `/models` | List available models from the current provider |
| `/providers` | List all built-in providers |
| `/thinking <n>` | Set reasoning effort 0–10 (0 = off) |
| `/reset` | Clear conversation history |
| `/exit` `/quit` `/q` | Exit |

Ctrl+C also exits cleanly.

## Features

- Streams responses with markdown→ANSI rendering (bold, inline code, code blocks with line numbers, tables, headings, bullets)
- Full filesystem access via tool calls: read files, write files, run shell commands, list directories
- Agentic loop — the model keeps calling tools until the task is done
- 40+ preconfigured providers (OpenAI, Z.AI, Groq, Together, Fireworks, Mistral, Cloudflare, Azure, and more)
- Persistent config at `~/.config/au/config.json` (plaintext JSON)
- Pinned status bar showing current model and thinking intensity
- Command autocomplete with Tab, Up/Down history navigation
- Thinking intensity control (0–10) for models that support `reasoning_effort`
- Single static binary, ~9 MB, ~8 MB RAM at idle
- Zero external dependencies except `golang.org/x/term`

## Config

Stored at `~/.config/au/config.json` (or `~/Library/Application Support/au/config.json` on macOS):

```json
{
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-...",
  "model": "gpt-4o",
  "thinking": 0
}
```

## Tools

The agent has access to four tools:

- `read_file` — read any file
- `write_file` — write or overwrite any file (creates parent directories)
- `run_command` — run any shell command via `sh -c` (60s timeout)
- `list_directory` — list directory contents with sizes

## Providers

Run `/providers` inside `au` for the full list. Highlights:

- OpenAI, Z.AI, Groq, Together AI, Fireworks, Mistral
- Cloudflare Workers AI (requires account ID)
- Azure OpenAI (requires endpoint + deployment)
- OpenRouter, DeepInfra, Perplexity, Cohere, Replicate, and more

## License

MIT
