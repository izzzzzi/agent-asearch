# SYSTEM_PROMPT_snippet.md — agent-asearch

Snippet for system prompts of coding agents (Claude Code, Codex, Cursor, etc.)
that instructs the agent how to use `asearch`.

```text
You have access to `asearch` — a multi-source search CLI for LLM agents.

All operational commands return JSON. Start with `asearch open --query "..." --source <sources>`.
Save the returned `sid`. Use `asearch results read -s SID --seq 1 --limit 20` for paginated
reads. Filter by source with `asearch results filter -s SID --source <source>`. Always close
sessions with `asearch session close -s SID`.

Sources available without configuration: hn (Hacker News), reddit, github, jina.
For web search, set any one of: TAVILY_API_KEY, EXA_API_KEY, BRAVE_API_KEY.
Check available backends with `asearch doctor`.

Prefer reading results in small chunks (--limit 20) to save tokens.
Use --raw for piping: asearch results read -s SID --raw | head -50.
Use next_commands from JSON responses to continue workflows.

Never echo API keys in responses. Keep returned sid values between calls.
Close sessions when done to free resources.

Quick reference:
  asearch open --query "..." --source web,hn,reddit
  asearch results read -s SID --seq 1 --limit 20
  asearch results filter -s SID --source reddit
  asearch session close -s SID
  asearch doctor
```
