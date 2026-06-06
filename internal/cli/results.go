package cli

import (
	"fmt"

	"github.com/izzzzzi/agent-asearch/internal/search"
	"github.com/spf13/cobra"
)

func newResultsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "results",
		Short:            "Read and filter search results",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeInvalidArgs(cmd, "results subcommand required", "asearch results read|filter")
		},
	}
	cmd.AddCommand(newResultsReadCommand())
	cmd.AddCommand(newResultsFilterCommand())
	return cmd
}

func newResultsReadCommand() *cobra.Command {
	var sid string
	var seq, limit int
	var raw bool

	cmd := &cobra.Command{
		Use:   "read -s SID [--seq 1 --limit 20]",
		Short: "Read paginated search results",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sid == "" {
				return writeError(cmd, "invalid_args", "sid required", "use -s SID from open response")
			}

			results, total, err := search.ReadResults(sid, seq, limit)
			if err != nil {
				return writeError(cmd, "read_error", err.Error(), "")
			}

			if raw {
				for _, r := range results {
					fmt.Fprintf(cmd.OutOrStdout(), "[%d] [%s] %s\n%s\n%s\n\n",
						r.Seq, r.Source, r.Title, r.URL, r.Snippet)
				}
				return nil
			}

			prevSeq := seq - limit
			if prevSeq < 1 {
				prevSeq = 1
			}

			return writeJSON(cmd, map[string]any{
				"ok":    true,
				"sid":   sid,
				"total": total,
				"seq":   seq,
				"limit": limit,
				"count": len(results),
				"results": map[string]any{
					"items": results,
				},
				"next_commands": map[string]string{
					"next":   fmt.Sprintf("asearch results read -s %s --seq %d --limit %d", sid, seq+limit, limit),
					"prev":   fmt.Sprintf("asearch results read -s %s --seq %d --limit %d", sid, prevSeq, limit),
					"filter": fmt.Sprintf("asearch results filter -s %s --source web", sid),
					"close":  fmt.Sprintf("asearch session close -s %s", sid),
				},
			})
		},
	}

	cmd.Flags().StringVarP(&sid, "sid", "s", "", "session id")
	_ = cmd.MarkFlagRequired("sid")
	cmd.Flags().IntVar(&seq, "seq", 1, "starting sequence number (1-indexed)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max results to return")
	cmd.Flags().BoolVar(&raw, "raw", false, "raw text output for piping")

	return cmd
}

func newResultsFilterCommand() *cobra.Command {
	var sid string
	var sourceFilter string

	cmd := &cobra.Command{
		Use:   "filter -s SID --source SOURCE",
		Short: "Filter results by source",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sid == "" {
				return writeError(cmd, "invalid_args", "sid required", "")
			}
			if sourceFilter == "" {
				return writeError(cmd, "invalid_args", "source required", "--source web|reddit|hn|github|youtube|twitter")
			}

			results, total, err := search.FilterResults(sid, search.Source(sourceFilter))
			if err != nil {
				return writeError(cmd, "filter_error", err.Error(), "")
			}

			return writeJSON(cmd, map[string]any{
				"ok":     true,
				"sid":    sid,
				"total":  total,
				"source": sourceFilter,
				"count":  len(results),
				"results": map[string]any{
					"items": results,
				},
			})
		},
	}

	cmd.Flags().StringVarP(&sid, "sid", "s", "", "session id")
	_ = cmd.MarkFlagRequired("sid")
	cmd.Flags().StringVar(&sourceFilter, "source", "", "filter by source")

	return cmd
}
