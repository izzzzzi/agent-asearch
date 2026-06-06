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
			"Start with asearch open --query \"your topic\" --source web,reddit,hn,github",
			"Keep returned sid and follow next_commands",
			"Use asearch results read -s SID --seq 1 --limit 20 for first page",
			"Use asearch results filter -s SID --source reddit to narrow by platform",
			"Use asearch doctor to verify available backends",
			"Always close sessions with asearch session close -s SID when finished",
			"Use --raw output for piping: asearch results read -s SID --raw | head -50",
		},
		Commands: map[string]string{
			"open":            "asearch open --query \"topic\" --source web,reddit,hn,github",
			"open_multi":      "asearch open --query \"topic\" --source tavily,web,reddit,hn,github,youtube",
			"open_tavily":     "asearch open --query \"topic\" --source tavily",
			"results_read":    "asearch results read -s SID --seq 1 --limit 20",
			"results_filter":  "asearch results filter -s SID --source reddit",
			"session_list":    "asearch session list",
			"session_close":   "asearch session close -s SID",
			"session_gc":      "asearch session gc",
			"doctor":          "asearch doctor",
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
Start with asearch open --query "your topic" --source searxng,web,hn,reddit,github,youtube
For zero-cost unlimited search, run: docker run -d -p 8080:8080 searxng/searxng && export ASEARCH_SEARXNG_URL=http://localhost:8080
For AI-optimized results, use --source tavily (needs TAVILY_API_KEY, free tier at tavily.com)
Save the returned sid, then use asearch results read -s SID --seq 1 --limit 20
Filter by source: asearch results filter -s SID --source reddit
Check available tools: asearch doctor | list sessions: asearch session list
Always close sessions: asearch session close -s SID

Sources (zero-config): searxng (self-host, docker), web (auto-delegate), hn, reddit, github, jina
Sources (needs API key): tavily (TAVILY_API_KEY), exa (EXA_API_KEY), brave (BRAVE_API_KEY)
Sources (needs tools): youtube (yt-dlp), twitter (twitter-cli)

Prefer reading results in small chunks (--limit 20) to save tokens.
Use --raw for piping: asearch results read -s SID --raw | head -50
Use next_commands from JSON responses to continue workflows.`,
	}
}
