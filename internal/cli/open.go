package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/izzzzzi/agent-asearch/internal/ids"
	"github.com/izzzzzi/agent-asearch/internal/search"
	"github.com/izzzzzi/agent-asearch/internal/session"
	"github.com/izzzzzi/agent-asearch/internal/state"
	"github.com/spf13/cobra"
)

func newOpenCommand() *cobra.Command {
	var name string
	var sources []string
	var limit int

	cmd := &cobra.Command{
		Use:   "open --query QUERY [--source web,reddit,hn...]",
		Short: "Start a search session with the given query and sources",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			query, _ := cmd.Flags().GetString("query")
			if query == "" {
				return writeError(cmd, "invalid_args", "query required", "use --query \"your search\"")
			}

			if err := state.EnsureDirs(); err != nil {
				return writeError(cmd, "state_error", err.Error(), "check ~/.asearch permissions")
			}

			srcs := parseSources(sources)
			if len(srcs) == 0 {
				srcs = []search.Source{search.SourceWeb}
			}

			sid, err := ids.New()
			if err != nil {
				return writeError(cmd, "internal_error", "failed to generate session id", "")
			}

			if name == "" {
				name = sanitizeName(query)
			}

			rec := session.Record{
				SID:       sid,
				Name:      name,
				Query:     query,
				Sources:   sourceStrings(srcs),
				CreatedAt: time.Now(),
				State:     "searching",
			}
			if err := session.Save(rec); err != nil {
				return writeError(cmd, "state_error", err.Error(), "")
			}

			req := search.SearchRequest{
				Query:   query,
				Sources: srcs,
				Limit:   limit,
				Timeout: 60 * time.Second,
			}

			result, searchErr := search.Search(req)
			if searchErr != nil {
				_ = session.UpdateState(sid, "error")
				return writeError(cmd, "search_error", searchErr.Error(), "run asearch doctor")
			}

			_ = session.UpdateTotal(sid, result.Total)
			_ = session.UpdateState(sid, "ready")

			if err := search.SaveResults(sid, result); err != nil {
				return writeError(cmd, "state_error", err.Error(), "")
			}

			srcNames := sourceStrings(srcs)
			return writeJSON(cmd, map[string]any{
				"ok":      true,
				"sid":     sid,
				"session": name,
				"query":   query,
				"sources": srcNames,
				"total":   result.Total,
				"next_commands": map[string]string{
					"read":      fmt.Sprintf("asearch results read -s %s --limit 20", sid),
					"read_seq":  fmt.Sprintf("asearch results read -s %s --seq 1 --limit 20", sid),
					"filter":    fmt.Sprintf("asearch results filter -s %s --source web", sid),
					"close":     fmt.Sprintf("asearch session close -s %s", sid),
					"list":      "asearch session list",
				},
			})
		},
	}

	cmd.Flags().String("query", "", "search query")
	cmd.Flags().StringSliceVar(&sources, "source", []string{"web"}, "search sources: web, reddit, hn, github, youtube, twitter")
	cmd.Flags().IntVar(&limit, "limit", 50, "max results per source")
	cmd.Flags().StringVarP(&name, "name", "n", "", "session name (default: derived from query)")

	return cmd
}

func parseSources(raw []string) []search.Source {
	seen := map[search.Source]bool{}
	var srcs []search.Source
	for _, r := range raw {
		for _, part := range strings.Split(r, ",") {
			part = strings.TrimSpace(strings.ToLower(part))
			var src search.Source
			switch part {
			case "web":
				src = search.SourceWeb
			case "reddit":
				src = search.SourceReddit
			case "twitter", "x":
				src = search.SourceTwitter
			case "hn", "hackernews", "hacker_news":
				src = search.SourceHN
			case "github", "gh":
				src = search.SourceGitHub
			case "youtube", "yt":
				src = search.SourceYouTube
			case "tavily":
				src = search.SourceTavily
			default:
				continue
			}
			if !seen[src] {
				seen[src] = true
				srcs = append(srcs, src)
			}
		}
	}
	return srcs
}

func sourceStrings(srcs []search.Source) []string {
	var s []string
	for _, src := range srcs {
		s = append(s, string(src))
	}
	return s
}

func sanitizeName(query string) string {
	name := strings.ToLower(query)
	name = strings.ReplaceAll(name, " ", "-")
	if len(name) > 32 {
		name = name[:32]
	}
	return name
}
