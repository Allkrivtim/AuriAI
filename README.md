# AuriAI

A self-hosted AI assistant framework in Go, built around a provider-agnostic core. The core is a black box that takes text and returns text — everything about *where* that text comes from and goes to is the job of interchangeable providers. Swap the LLM backend, the storage engine, or the I/O channel without touching the agent logic.

The assistant currently runs as a Telegram bot backed by any OpenAI-compatible LLM, with web search, long-term memory, and note-taking exposed to the model as tools.

## Design

The project follows a ports-and-adapters (hexagonal) architecture. The `core` package defines interfaces and the agent loop; it depends on nothing concrete. Every external concern — the LLM, the database, the messaging channel, each tool — is an adapter that satisfies a core interface and is wired together in `main.go`.

```
        I/O providers                 CORE (agent logic)              LLM provider
     ┌──────────────────┐      ┌──────────────────────────┐      ┌─────────────────┐
     │  telegram        │─────▶│  Engine.Handle(in) → out │      │  openai-compat  │
     │  (in / out)      │◀─────│                          │─────▶│  (any base URL) │
     └──────────────────┘      │  · agent loop            │◀─────└─────────────────┘
                               │  · history / sessions    │
                               │  · tool dispatch         │       ┌─────────────────┐
                               │  · dynamic system prompt │──────▶│  SQLite store   │
                               └──────────────────────────┘       └─────────────────┘
                                          │
                                   ┌──────┴───────┐
                                   │  tools       │
                                   │  websearch   │
                                   │  memory      │
                                   │  notes       │
                                   └──────────────┘
```

Two boundaries surround the core. On the left, I/O providers call `Engine.Handle` — they only marshal an inbound message and send back a reply, knowing nothing about the LLM or the database. On the right, the core calls out through the `LLM`, `Store`, `MemoryStore`, and `Tool` interfaces — it depends on the contracts, never the implementations. This is what makes each piece replaceable in a single line of wiring.

The heart of the core is an **agent loop**. A user message doesn't map to a single LLM call: the model may respond with tool calls instead of text, in which case the core executes each tool, appends the results to the conversation history, and calls the model again — repeating until the model returns a plain text answer or a step limit is reached.

## Features

- **Provider-agnostic LLM** — works with any OpenAI-compatible endpoint (OpenRouter, local inference servers, the OpenAI API itself) via a configurable base URL.
- **Tool calling / agent loop** — the model can invoke tools and reason over their results across multiple turns within a single user request.
- **Web search** — a built-in tool backed by a search API, giving the assistant access to current information.
- **Long-term memory** — the assistant can save, retrieve, and delete keyed memories that persist across conversations.
- **Notes** — short-lived notes with a time-to-live, optional pin (surfaced directly in the system prompt), and a notification flag reserved for upcoming scheduled wake-ups.
- **Dynamic system prompt** — the current date/time and pinned notes are injected into the system prompt on every turn, so the model always has fresh context.
- **Persistent history** — conversations, memories, and notes are stored in SQLite; each Telegram chat maps to its own isolated session.
- **Telegram provider** — long-polling bot with typing indicators, automatic splitting of long replies, per-chat sessions, and group-mention handling.

## Tech stack

- **Language:** Go 1.26
- **LLM client:** [`openai-go`](https://github.com/openai/openai-go) (OpenAI-compatible)
- **Storage:** SQLite via [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO) with named queries loaded through [`dotsql`](https://github.com/qustavo/dotsql)
- **Telegram:** [`go-telegram-bot-api`](https://github.com/go-telegram-bot-api/telegram-bot-api)
- **Config:** [`godotenv`](https://github.com/joho/godotenv)

## Project layout

```
cmd/server/            entry point — wires all adapters into the engine
internal/
  core/                the hexagon: interfaces, types, agent loop, registry
    ports.go           Engine, LLM, Store, MemoryStore, Tool, ToolRegistry
    engine.go          agent loop + dynamic system prompt
    types.go           domain types (Message, ToolCall, Memory, Note, ...)
    registry.go        in-memory tool registry
  adapters/
    llm/openai/        OpenAI-compatible LLM adapter
    store/sqlite/      SQLite persistence (schema + queries embedded)
    telegram/          Telegram I/O provider (bot + handlers)
    tools/
      websearch/       web search tool
      memory/          long-term memory tool
      notes/           notes tool
  prompts/
    base.md            base system prompt
```

## Getting started

### Prerequisites

- Go 1.26 or newer
- A Telegram bot token (from [@BotFather](https://t.me/BotFather))
- An OpenAI-compatible LLM endpoint + API key
- A search provider API key (for the web search tool)

### Configuration

Copy the example environment file and fill in your values:

```bash
cp example.env .env
```

```ini
# I/O Providers
TG_BOT_TOKEN=your-telegram-bot-token

# LLM Provider (any OpenAI-compatible endpoint)
LLM_URL=https://your-llm-provider/v1
LLM_API_KEY=your-llm-api-key
LLM_MODEL=your-model-name

# Tools
SEARCH_PROVIDER_API_KEY=your-search-api-key
```

### Run

```bash
# create the storage directory (SQLite file lives here)
mkdir -p storage

# build and run
go build -o ./build/server ./cmd/server
./build/server
```

The bot connects to Telegram via long polling. Message it directly, or mention it in a group.

## How it works

A message flows through the system like this:

1. The Telegram provider receives an update and calls `Engine.Handle` with an inbound message carrying the session ID and text.
2. The core stores the user message and enters the agent loop.
3. Each iteration loads the recent history, builds a system prompt (base prompt + current time + pinned notes), and calls the LLM with the available tool specs.
4. If the model returns plain text, it's stored and returned to the provider, which replies to the user.
5. If the model returns tool calls, the core executes each one, stores the results in history, and loops again so the model can reason over them.

Tool errors are fed back to the model as text rather than crashing the request — the assistant can retry or explain the failure, instead of the user just seeing an error.

## Roadmap

The framework is built in stages toward an increasingly autonomous assistant.

- [x] Provider-agnostic core with agent loop
- [x] OpenAI-compatible LLM adapter
- [x] SQLite persistence with per-chat sessions
- [x] Telegram I/O provider
- [x] Web search tool
- [x] Long-term memory + notes tools
- [x] Dynamic system prompt (time + pinned notes)
- [ ] **Heartbeat** — scheduled wake-ups so the assistant can act proactively (e.g. surface expired notes, act on notifications) rather than only reacting to messages
- [ ] Additional I/O providers (REST/HTTP, CLI)
- [ ] Streaming responses
- [ ] Expiry handling for notes (TTL-based cleanup)
- [ ] MCP support for dynamically connecting external tools

## Status

Early and under active development — interfaces and internals may change between commits.

## License

_No license specified yet._