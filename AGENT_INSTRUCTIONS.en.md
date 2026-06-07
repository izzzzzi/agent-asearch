# AGENT_INSTRUCTIONS — agent-asearch

Instructions for LLM agents using `asearch`.

→ Short version: `asearch prompt`
→ Check availability: `asearch doctor`

---

## Quick Start (30 seconds to first result)

```bash
# Nothing to install — HN works immediately
asearch open --query "claude code plugins" --source hn
asearch results read -s SID --seq 1 --limit 20
asearch session close -s SID
```

---

## Provider Setup

### Zero-config — works immediately

| Provider | Command |
|-----------|---------|
| **hn** | `asearch open --query "..." --source hn` |
| **github** | `asearch open --query "..." --source github` (needs `gh` on PATH) |
| **code** | `asearch open --query "func" --source code` (GH code search) |
| **jina** | `asearch open --query "..." --source jina` (URL → markdown) |
| **reddit** | Save cookies: `~/.asearch/reddit-cookies.txt` (Netscape from browser) |

### Manage API keys (persistent, no env vars)

```bash
asearch config set tavily "tvly-..."   # save a key
asearch config set exa "..."           # any provider
asearch config show                     # show all keys (masked)
```

### Self-hosted — your own server, no limits

```bash
# SearXNG — aggregates Google, Bing, DDG, Wikipedia + 70 engines
docker run -d -p 8080:8080 searxng/searxng
export ASEARCH_SEARXNG_URL=http://localhost:8080
asearch open --query "..." --source searxng
```

### API keys — pick any one, all have free tiers

| Provider | Env var / config key | Where to get | Free |
|-----------|----------------------|--------------|:----:|
| **Tavily** | `tavily` | tavily.com | ✅ |
| **Perplexity** | `perplexity` | docs.perplexity.ai | ✅ |
| **Exa** | `exa` | exa.ai | ✅ |
| **Brave** | `brave` | brave.com/search/api | 2000/mo |
| **Serper** | `serper` | serper.dev | 2500/mo |
| **SerpAPI** | `serpapi` | serpapi.com | 100/mo |
| **You.com** | `you` | you.com/api | ✅ |
| **Firecrawl** | `firecrawl` | firecrawl.dev | 500/mo |
| **Parallel** | `parallel` | parallel.ai | ✅ |
| **Twitter** | `TWITTER_BEARER_TOKEN` | developer.twitter.com | 500k/month |

### Needs cookies from browser

| Provider | Setup |
|----------|-------|
| **youtube** | Save cookies to `~/.asearch/youtube-cookies.txt` (Netscape format) |
| **reddit** | Save cookies to `~/.asearch/reddit-cookies.txt` |

### Needs install

| Provider | Setup |
|----------|-------|
| **twitter** | Bearer Token from developer.twitter.com → `asearch config set twitter "AAAA..."` |

---

## Commands

| Command | Description |
|---------|-------------|
| `asearch doctor` | Check all 18 providers — what's ready, what's not |
| `asearch open --query Q --source SRC` | Start a search session |
| `asearch open --query Q -x` | Cross-reference (tool by author) |
| `asearch open --query Q --synth` | Inline results + agent synthesis prompt |
| `asearch results read -s SID --seq N --limit M` | Paginated read |
| `asearch results filter -s SID --source SRC` | Filter by source |
| `asearch session list` | List sessions |
| `asearch session close -s SID` | Close session |
| `asearch config set|get|show` | Manage keys |
| `asearch update` | Self-update |
| `asearch reddit sub NAME -l hot` | Browse subreddit |
| `asearch reddit read /r/...` | Read post + comments |
| `asearch reddit info NAME` | Subreddit info |
| `asearch completion zsh` | Shell completion |

---

## Web delegation chain

`--source web` tries providers in order, first available wins:

```
SearXNG → DDG HTML → Wikipedia → Bing HTML → Tavily → Perplexity → Exa → Brave → Serper → ...
```

---

## JSON contract

### Success (`open`)
```json
{
  "ok": true, "sid": "a1b2c3d4", "total": 42,
  "next_commands": {
    "read": "asearch results read -s a1b2c3d4 --limit 20",
    "filter": "asearch results filter -s a1b2c3d4 --source reddit",
    "close": "asearch session close -s a1b2c3d4"
  }
}
```

### Error
```json
{"ok":false,"code":"invalid_args","message":"query required","hint":"use --query \"...\""}
```

---

## Token economy

1. `open` → check `total`, assess volume
2. `results filter --source X` → narrow down
3. `results read --seq 1 --limit 20` → small pages
4. `--raw` for grep/head piping

---

## Example workflow

```bash
# 1. Check what's available
asearch doctor

# 2. Set an API key
asearch config set tavily "tvly-..."

# 3. Search
asearch open --query "best AI coding tools 2026" --source web,hn,github

# 4. Read
asearch results read -s SID --seq 1 --limit 20

# 5. Filter
asearch results filter -s SID --source hn

# 6. Done
asearch session close -s SID
```
