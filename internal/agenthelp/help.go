package agenthelp

type HelpPayload struct {
	OK                 bool              `json:"ok"`
	Tool               string            `json:"tool"`
	Audience           string            `json:"audience"`
	Kind               string            `json:"kind"`
	CommandGroup       string            `json:"command_group,omitempty"`
	AgentPromptCommand string            `json:"agent_prompt_command"`
	Docs               []string          `json:"docs,omitempty"`
	Workflow           []string          `json:"workflow"`
	Commands           map[string]string `json:"commands"`
}

type PromptPayload struct {
	OK       bool   `json:"ok"`
	Tool     string `json:"tool"`
	Audience string `json:"audience"`
	Kind     string `json:"kind"`
	Prompt   string `json:"prompt"`
}

func RootHelp() HelpPayload {
	return HelpPayload{
		OK: true, Tool: "asearch", Audience: "llm_agent", Kind: "agent_help",
		AgentPromptCommand: "asearch prompt",
		Docs:               []string{"AGENT_INSTRUCTIONS.md", "README.md"},
		Workflow: []string{
			"Start with asearch open --query \"your topic\" --source web,reddit,hn,github,youtube",
			"Keep returned sid and follow next_commands",
			"Use asearch results read -s SID --seq 1 --limit 20 for first page",
			"Use asearch results filter -s SID --source reddit to narrow by platform",
			"Use asearch doctor to verify available backends",
			"Always close sessions with asearch session close -s SID when finished",
			"Use --raw output for piping: asearch results read -s SID --raw | head -50",
		},
		Commands: map[string]string{
			"open":            "asearch open --query \"topic\" --source web,reddit,hn,github",
			"open_multi":      "asearch open --query \"topic\" --source web,hn,reddit,github,youtube,code",
			"open_code":       "asearch open --query \"func\" --source code",
			"results_read":    "asearch results read -s SID --seq 1 --limit 20",
			"results_filter":  "asearch results filter -s SID --source reddit",
			"session_list":    "asearch session list",
			"session_close":   "asearch session close -s SID",
			"session_gc":      "asearch session gc",
			"doctor":          "asearch doctor",
			"config_show":     "asearch config show",
			"config_set":      "asearch config set <key> <value>",
			"prompt":          "asearch prompt",
			"version":         "asearch version",
		},
	}
}

func GroupHelp(name string) (HelpPayload, bool) {
	groups := map[string]HelpPayload{
		"results": {
			OK: true, Tool: "asearch", Audience: "llm_agent", Kind: "agent_help",
			CommandGroup: "results", AgentPromptCommand: "asearch prompt",
			Workflow: []string{
				"Use results read with --limit 20 to keep token usage low",
				"Use results filter to narrow by source before reading",
				"Use next_commands to paginate forward",
				"Use --raw for grep/head/pipe workflows",
			},
			Commands: map[string]string{
				"read":   "asearch results read -s SID --seq 1 --limit 20",
				"filter": "asearch results filter -s SID --source web",
			},
		},
		"session": {
			OK: true, Tool: "asearch", Audience: "llm_agent", Kind: "agent_help",
			CommandGroup: "session", AgentPromptCommand: "asearch prompt",
			Workflow: []string{
				"Use session list to discover active sessions",
				"Always close sessions after use",
				"Use session gc to clean old closed sessions",
			},
			Commands: map[string]string{
				"list":  "asearch session list",
				"close": "asearch session close -s SID",
				"gc":    "asearch session gc",
			},
		},
	}
	payload, ok := groups[name]
	return payload, ok
}

func Prompt() PromptPayload {
	return PromptPayload{
		OK: true, Tool: "asearch", Audience: "llm_agent", Kind: "agent_prompt",
		Prompt: `You are using asearch, a multi-source search CLI for LLM agents.
All operational commands return JSON.
Start with asearch open --query "your topic" --source web,hn,reddit,github,youtube,code
Save the returned sid, then use asearch results read -s SID --seq 1 --limit 20
Filter by source: asearch results filter -s SID --source reddit
Check available tools: asearch doctor | list sessions: asearch session list
Always close sessions: asearch session close -s SID

Zero-config sources: web (DDG/Wikipedia/Bing), hn, reddit (cookies), github, jina
Self-hosted: searxng (docker run searxng/searxng + ASEARCH_SEARXNG_URL)
API keys (any one): tavily, exa, brave, serper, serpapi, you, firecrawl, parallel, perplexity
  Save with: asearch config set <provider> <key>
Needs cookies: youtube (~/.asearch/youtube-cookies.txt), reddit (~/.asearch/reddit-cookies.txt)
Needs install: twitter (pipx install twitter-cli)`,
	}
}
