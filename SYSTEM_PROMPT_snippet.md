# SYSTEM_PROMPT_snippet.md — agent-asearch

Snippet for system prompts of coding agents (Claude Code, Codex, Cursor, etc.)
that instructs the agent how to use `asearch`.

```text
You have access to `asearch` — a multi-source search CLI for LLM agents.

All operational commands return JSON. Start with `asearch open --query "..." --source <sources>`.
Save the returned `sid`. Use `asearch results read -s SID --seq 1 --limit 20` for paginated
reads. Filter by source with `asearch results filter -s SID --source <source>`. Always close
sessions with `asearch session close -s SID`.

Sources available without configuration: hn, reddit, github, jina.
For YouTube: save browser cookies to ~/.asearch/youtube-cookies.txt.
For web search with API keys: asearch config set tavily|exa|brave <key>.
For zero-cost self-hosted search: docker run searxng/searxng + ASEARCH_SEARXNG_URL.
Check available backends with `asearch doctor`.

Prefer reading results in small chunks (--limit 20) to save tokens.
Use --raw for piping: asearch results read -s SID --raw | head -50.
Use next_commands from JSON responses to continue workflows.
Use --cross-ref (-x) to cross-reference results (e.g., "tool by author").
Use --synth to get inline results + next_commands.synth — a synthesis prompt for the agent to analyze and produce structured JSON output.

Never echo API keys in responses. Keep returned sid values between calls.
Close sessions when done to free resources.

Quick reference:
  asearch open --query "..." --source searxng,web,hn,reddit
  asearch results read -s SID --seq 1 --limit 20
  asearch results filter -s SID --source reddit
  asearch session close -s SID
  asearch doctor
```
