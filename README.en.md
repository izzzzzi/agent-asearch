# asearch

[![npm](https://img.shields.io/npm/v/agent-asearch)](https://www.npmjs.com/package/agent-asearch)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Language: [Русский](README.md) | English

Search CLI for LLM agents. One command, 18 sources.

`asearch` searches the web, Hacker News, Reddit, GitHub, YouTube, X/Twitter, code, and 9 API providers simultaneously. Keeps agent context clean: compact metadata first, then paginated reads of only what you need. One Go binary, single dependency — Cobra.

![asearch architecture](docs/asearch-architecture.png)

## Quick Start

```bash
npm i -g agent-asearch

# Zero-config — works immediately, nothing to install
asearch open --query "claude code plugins" --source hn,reddit

# Web search — DDG + Wikipedia + Bing, no keys needed
asearch open --query "claude code plugins" --source web

# Code search — gh search code
asearch open --query "error handling golang" --source code

# API keys persist in config (no env vars)
asearch config set tavily "tvly-..."   # tavily.com
asearch config set exa "..."           # exa.ai

# With a key — full web search
asearch open --query "claude code plugins" --source web,hn,reddit,github
```

`open` returns a session id and `next_commands`:

```json
{
  "ok": true,
  "sid": "a1b2c3d4",
  "session": "claude-code-plugins",
  "sources": ["web", "hn", "reddit"],
  "total": 42,
  "next_commands": {
    "read": "asearch results read -s a1b2c3d4 --limit 20",
    "filter": "asearch results filter -s a1b2c3d4 --source reddit",
    "close": "asearch session close -s a1b2c3d4"
  }
}
```

Continue through the results API:

```bash
asearch results read -s a1b2c3d4 --seq 1 --limit 20
asearch results filter -s a1b2c3d4 --source reddit
asearch results read -s a1b2c3d4 --raw | head -50
asearch session close -s a1b2c3d4
```

## What open does

`asearch open`:

- parses `--source` and selects available search backends;
- for `web`, tries in order: SearXNG → DDG HTML → Wikipedia → Bing HTML → API providers;
- runs search in parallel across all selected sources;
- saves results locally for paginated reading;
- returns `sid`, `total`, and `next_commands`.

## Sources

| Source | Out-of-box | Requires |
|--------|:----------:|----------|
| **web** | ✅ | DDG → Wikipedia → Bing HTML scrapers (no key) |
| **hn** | ✅ | Algolia HN Search API (free, no key) |
| **reddit** | ✅ cookies | Browser cookies → ~/.asearch/reddit-cookies.txt |
| **github** | ✅ | `gh` CLI |
| **code** | ✅ | GitHub code search via `gh` |
| **youtube** | ✅ cookies | Browser cookies → ~/.asearch/youtube-cookies.txt |
| **jina** | ✅ | URL → markdown reader (jina.ai, no key) |
| **searxng** | 🐳 | `docker run searxng/searxng` + ASEARCH_SEARXNG_URL |
| **tavily** | 🔑 | `asearch config set tavily ...` — AI answers |
| **exa** | 🔑 | `asearch config set exa ...` — neural search |
| **brave** | 🔑 | `asearch config set brave ...` — 35B index |
| **serper** | 🔑 | Google SERP (2500 free/month) |
| **serpapi** | 🔑 | 40+ search engines |
| **perplexity** | 🔑 | AI answers with citations |
| **you** | 🔑 | You.com search |
| **firecrawl** | 🔑 | JS-rendered web scraping |
| **parallel** | 🔑 | Parallel.ai search |
| **twitter** | ✅ built-in | Guest API (anonymous) or Bearer Token: `asearch config set twitter "..."` |

## Optional OpenClaw X/Twitter handoff

Keep `asearch` as the discovery step. If an agent needs account-scoped X/Twitter actions after search results, install [TweetClaw](https://github.com/Xquik-dev/tweetclaw) separately:

```bash
openclaw plugins install npm:@xquik/tweetclaw
```

Use it for reviewed workflows such as search tweet replies, follower export, user lookup, media upload and download, monitors, webhooks, direct messages, giveaway draws, and approval-gated post tweets or replies. Do not share `asearch` sessions or browser cookies with TweetClaw; configure the Xquik API key through the OpenClaw plugin settings.

## Commands

- `asearch open --query Q --source SRC` — start a search session.
- `asearch open --query Q -x` / `--cross-ref` — cross-reference results (tool by author).
- `asearch open --query Q --synth` — inline results + agent synthesis prompt.
- `asearch results read -s SID --seq N --limit M` — paginated read.
- `asearch results filter -s SID --source SRC` — filter by source.
- `asearch session list|close|gc` — session management.
- `asearch config set|get|show` — API key management.
- `asearch reddit sub|read|info` — browse Reddit.
- `asearch doctor` — check available backends.
- `asearch update` — self-update.
- `asearch completion bash|zsh|fish` — shell completion.
- `asearch prompt` — LLM agent instructions.
- `asearch version` — version info.

## Token economy

Check metadata first, then read only what you need:

```bash
asearch open --query "rust async benchmarks" --source web,hn
# → {"ok":true,"sid":"...","total":42,...}

asearch results read -s SID --seq 1 --limit 10
asearch results read -s SID --seq 11 --limit 10
asearch results filter -s SID --source hn
```

Use `--raw` for piping:

```bash
asearch results read -s SID --raw | grep "tokio"
```

## Agent CLI examples

Paste this into Codex, Claude Code, OpenCode, or any terminal agent:

```text
Use `asearch` for search tasks.

Check available backends first:
asearch doctor

For quick search without keys, use hn and reddit:
asearch open --query "your topic" --source hn,reddit

For full web search, set a key:
export TAVILY_API_KEY="tvly-..."
asearch open --query "your topic" --source web,hn,reddit,github

Save the returned sid. Read results in small pages:
asearch results read -s SID --seq 1 --limit 20

Filter by source before reading:
asearch results filter -s SID --source reddit

For piping, use --raw:
asearch results read -s SID --raw | head -50

Always close the session:
asearch session close -s SID
```

Short versions for popular CLIs:

```text
Codex: Use `asearch` for search. Start with `asearch doctor`, then `asearch open --query "..." --source web,hn`. Keep sid, read via `asearch results read -s SID --seq 1 --limit 20`, filter via `asearch results filter`. Close session.
```

```text
Claude Code: Before search, install/run `asearch`. Use `asearch open --query "..." --source web,hn,reddit`, keep sid, paginate through `asearch results read`, filter by source, close session.
```

```text
OpenCode: Use `asearch open`, then `asearch results read/filter` with returned sid. Don't mix sids across sessions. Run `asearch doctor` to check backends.
```

## Security

- Search queries are not written to audit logs.
- Sessions and results stored locally in `~/.asearch/`.
- GitHub token read from `GITHUB_TOKEN` or `GH_TOKEN`.
- API keys accepted only via env vars. No `--api-key` flags.
- Jina Reader works without a key (rate-limited); key lifts limits.

## Pros

- One command to search across 10 sources.
- Stable JSON responses for agent parsing.
- Token-efficient: metadata first, paginated reads.
- Session-based workflow with `sid` and `next_commands` — same as `assh` and `aget`.
- 5 sources work without API keys.
- Pluggable backend architecture — each source is a separate Go struct.
- Go single binary — no Python, no Node, no Docker.

## Limitations

- `web` without an API key shows setup instructions instead of results.
- Reddit public JSON may rate-limit without cookies (save cookies from browser to ~/.asearch/reddit-cookies.txt).
- Twitter search uses guest API (anonymous) or optional X API Bearer Token for higher reliability.
- YouTube works via browser cookies (save to ~/.asearch/youtube-cookies.txt).

## Manual install

`npm i -g agent-asearch` installs a wrapper that downloads the matching Go binary from GitHub Releases. Archives can be downloaded manually:

```text
https://github.com/izzzzzi/agent-asearch/releases
```

Or build from source:

```bash
git clone https://github.com/izzzzzi/agent-asearch
cd agent-asearch
go build -o asearch ./cmd/asearch
```

## License

MIT
