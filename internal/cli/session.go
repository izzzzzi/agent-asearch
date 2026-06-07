package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/izzzzzi/agent-asearch/internal/search"
	"github.com/izzzzzi/agent-asearch/internal/session"
	"github.com/izzzzzi/agent-asearch/internal/state"
	"github.com/spf13/cobra"
)

func newSessionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "session",
		Short:            "Manage search sessions",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeInvalidArgs(cmd, "session subcommand required", "asearch session list|close|gc")
		},
	}
	cmd.AddCommand(newSessionListCommand())
	cmd.AddCommand(newSessionCloseCommand())
	cmd.AddCommand(newSessionGcCommand())
	return cmd
}

func newSessionListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all search sessions",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := state.EnsureDirs(); err != nil {
				return writeError(cmd, "state_error", err.Error(), "")
			}
			sessions := session.List()
			if sessions == nil {
				sessions = []session.Record{}
			}
			return writeJSON(cmd, map[string]any{
				"ok":       true,
				"sessions": sessions,
				"count":    len(sessions),
			})
		},
	}
}

func newSessionCloseCommand() *cobra.Command {
	var sid string
	cmd := &cobra.Command{
		Use:   "close -s SID",
		Short: "Close a search session",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sid == "" {
				return writeError(cmd, "invalid_args", "sid required", "-s SID")
			}
			if err := session.Remove(sid); err != nil {
				return writeError(cmd, "session_error", err.Error(), "")
			}
			_ = session.Unlock(sid)
			return writeJSON(cmd, map[string]any{
				"ok":    true,
				"sid":   sid,
				"state": "closed",
			})
		},
	}
	cmd.Flags().StringVarP(&sid, "sid", "s", "", "session id")
	_ = cmd.MarkFlagRequired("sid")
	return cmd
}

func newSessionGcCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gc",
		Short: "Garbage-collect old closed sessions",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			count := 0
			sessions := session.List()
			for _, s := range sessions {
				if s.State == "closed" {
					_ = session.Remove(s.SID)
					count++
				}
			}
			return writeJSON(cmd, map[string]any{
				"ok":      true,
				"cleaned": count,
				"total":   len(sessions),
			})
		},
	}
}

func newPromptCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "prompt",
		Short: "Print agent usage instructions",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := `You are using asearch — a multi-source search CLI for LLM agents (16 providers).
All operational commands return JSON with sid + next_commands.

== QUICK START ==
  asearch doctor                              # check what's available
  asearch open --query "topic" --source hn   # zero-config, works immediately
  asearch results read -s SID --seq 1 --limit 20
  asearch session close -s SID

== INSTALL & ACTIVATE EACH PROVIDER ==

Zero-config (works immediately, nothing to install):
  hn       — asearch open --query "..." --source hn
  reddit   — save cookies: ~/.asearch/reddit-cookies.txt
            Browse: asearch reddit sub NAME -l hot
            Read:   asearch reddit read /r/.../comments/ID
            Info:   asearch reddit info NAME
  github   — gh CLI already on PATH: asearch open --query "..." --source github
  jina     — asearch open --query "..." --source jina  (URL-to-markdown reader)
  youtube  — pipx install yt-dlp && asearch open --query "..." --source youtube

Self-hosted (unlimited, zero cost, 30 seconds):
  searxng  — docker run -d -p 8080:8080 searxng/searxng
             export ASEARCH_SEARXNG_URL=http://localhost:8080
             asearch open --query "..." --source searxng

API keys (pick any one, free tiers available):
  tavily      — export TAVILY_API_KEY="tvly-..."       (free at tavily.com)
  perplexity  — export PERPLEXITY_API_KEY="pplx-..."    (free at docs.perplexity.ai)
  exa         — export EXA_API_KEY="..."                (free at exa.ai)
  brave       — export BRAVE_API_KEY="BSA..."           (2000 free/mo at brave.com/search/api)
  serper      — export SERPER_API_KEY="..."             (2500 free/mo at serper.dev)
  serpapi     — export SERPAPI_API_KEY="..."            (100 free/mo at serpapi.com)
  you         — export YOU_API_KEY="..."                (free at you.com/api)
  firecrawl   — export FIRECRAWL_API_KEY="fc-..."       (500 free/mo at firecrawl.dev)
  parallel    — export PARALLEL_API_KEY="..."           (free at parallel.ai)

  After setting any key, --source web auto-delegates:
  SearXNG → Tavily → Perplexity → Exa → Brave → Serper → SerpAPI → You → Firecrawl → Parallel

Browser auth (one-time):
  twitter — pipx install twitter-cli && twitter login  (opens browser, saves cookies)
            Then: asearch open --query "..." --source twitter

== TOKEN ECONOMY ==
  Prefer --limit 20 for compact JSON.
  Filter before reading: asearch results filter -s SID --source reddit
  Pipe with --raw: asearch results read -s SID --raw | head -50
  Always close sessions: asearch session close -s SID

== DEDUPLICATION ==
  Cross-source duplicates are automatically removed by normalized URL.
  First occurrence wins — earlier backends in the chain have priority.
  Same URL on Reddit + HN + web = 1 result, first source listed.`
			fmt.Fprintln(cmd.OutOrStdout(), prompt)
			return nil
		},
	}
}

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check available search backends and tools",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			type check struct {
				Name      string `json:"name"`
				Available bool   `json:"available"`
				Tool      string `json:"tool,omitempty"`
				Note      string `json:"note,omitempty"`
			}
			checks := []check{
				{Name: "searxng", Available: os.Getenv("ASEARCH_SEARXNG_URL") != "", Tool: "searxng", Note: "Self-hosted meta-search, zero cost, no limits; docker run searxng/searxng"},
				{Name: "tavily", Available: tavilyAvailable(), Tool: "tavily", Note: "AI-optimized search; set TAVILY_API_KEY (free tier at tavily.com)"},
				{Name: "perplexity", Available: os.Getenv("PERPLEXITY_API_KEY") != "", Tool: "perplexity", Note: "AI answers with citations; set PERPLEXITY_API_KEY"},
				{Name: "exa", Available: os.Getenv("EXA_API_KEY") != "", Tool: "exa", Note: "Neural/semantic search; set EXA_API_KEY (exa.ai)"},
				{Name: "brave", Available: os.Getenv("BRAVE_API_KEY") != "", Tool: "brave", Note: "35B-page index, 2000 free/month; set BRAVE_API_KEY (brave.com/search/api)"},
				{Name: "serper", Available: os.Getenv("SERPER_API_KEY") != "", Tool: "serper", Note: "Google SERP; set SERPER_API_KEY (2500 free/month at serper.dev)"},
				{Name: "serpapi", Available: os.Getenv("SERPAPI_API_KEY") != "", Tool: "serpapi", Note: "40+ search engines; set SERPAPI_API_KEY (100 free/month at serpapi.com)"},
				{Name: "you", Available: os.Getenv("YOU_API_KEY") != "", Tool: "you.com", Note: "You.com web search; set YOU_API_KEY"},
				{Name: "firecrawl", Available: os.Getenv("FIRECRAWL_API_KEY") != "", Tool: "firecrawl", Note: "JS-rendered scraping; set FIRECRAWL_API_KEY (500 free/month)"},
				{Name: "parallel", Available: os.Getenv("PARALLEL_API_KEY") != "", Tool: "parallel", Note: "Parallel.ai search; set PARALLEL_API_KEY"},
				{Name: "jina", Available: true, Tool: "jina", Note: "URL-to-markdown reader; set JINA_API_KEY for higher rate limits (jina.ai)"},
				{Name: "web", Available: true, Note: "DuckDuckGo Lite (auto-delegates to tavily/brave/exa if keys set)"},
				{Name: "reddit", Available: true, Note: "Reddit JSON API (browse/read/info); needs cookies from browser"},
				{Name: "hn", Available: true, Note: "Algolia API"},
				{Name: "github", Available: toolAvailable("gh"), Tool: "gh", Note: "GitHub CLI"},
				{Name: "youtube", Available: toolAvailable("yt-dlp"), Tool: "yt-dlp", Note: "install with: brew install yt-dlp"},
				{Name: "twitter", Available: toolAvailable("twitter"), Tool: "twitter-cli", Note: "install with: pipx install twitter-cli"},
			}

			if err := state.EnsureDirs(); err != nil {
				return writeError(cmd, "state_error", err.Error(), "check ~/.asearch permissions")
			}

			allOK := true
			for _, c := range checks {
				if !c.Available && c.Tool != "" {
					allOK = false
				}
			}

			return writeJSON(cmd, map[string]any{
				"ok":        allOK,
				"checks":    checks,
				"state_dir": state.StateDir(),
				"hint":      "Run asearch doctor to verify installation",
			})
		},
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeJSON(cmd, map[string]any{
				"ok":      true,
				"tool":    "asearch",
				"version": "0.1.0",
				"go":      "1.24",
			})
		},
	}
}

func toolAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Sub-command for per-source search (web, reddit, hn, github, etc.)
func newWebSearchCommand() *cobra.Command {
	var query string
	var limit int
	cmd := &cobra.Command{
		Use:   "search --query QUERY",
		Short: "Direct web search (no session)",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if query == "" {
				return writeError(cmd, "invalid_args", "query required", "--query TEXT")
			}
			backend := &search.WebBackend{}
			results, err := backend.Search(query, limit)
			if err != nil {
				return writeError(cmd, "search_error", err.Error(), "")
			}
			return writeJSON(cmd, map[string]any{
				"ok":      true,
				"source":  "web",
				"query":   query,
				"results": results,
				"count":   len(results),
			})
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "search query")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "result limit")
	return cmd
}

func tavilyAvailable() bool {
	return os.Getenv("TAVILY_API_KEY") != ""
}
