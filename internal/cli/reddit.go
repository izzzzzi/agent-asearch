package cli

import (
	"time"

	"github.com/izzzzzi/agent-asearch/internal/search"
	"github.com/spf13/cobra"
)

func newRedditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reddit",
		Short: "Browse Reddit — subreddits, posts, comments",
	}
	cmd.AddCommand(newRedditSubCommand())
	cmd.AddCommand(newRedditReadCommand())
	cmd.AddCommand(newRedditInfoCommand())
	return cmd
}

func newRedditSubCommand() *cobra.Command {
	var listing string
	var limit int
	cmd := &cobra.Command{
		Use:   "sub NAME [--hot|--new|--top|--rising]",
		Short: "Browse subreddit posts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rc := search.NewRedditClient()
			posts, err := rc.SubredditPosts(args[0], listing, limit)
			if err != nil {
				return writeError(cmd, "reddit_error", err.Error(), "")
			}

			type item struct {
				ID       string  `json:"id"`
				Title    string  `json:"title"`
				URL      string  `json:"url"`
				Score    int     `json:"score"`
				Ratio    float64 `json:"upvote_ratio"`
				Comments int     `json:"comments"`
				Author   string  `json:"author"`
				Date     string  `json:"date"`
				Flair    string  `json:"flair,omitempty"`
			}
			var items []item
			for _, p := range posts {
				if p.Stickied && listing == "hot" {
					continue
				} // skip stickies
				items = append(items, item{
					ID: p.ID, Title: p.Title,
					URL:   "https://www.reddit.com" + p.Permalink,
					Score: p.Score, Ratio: p.UpvoteRatio,
					Comments: p.NumComments, Author: p.Author,
					Date:  time.Unix(int64(p.Created), 0).Format("2006-01-02"),
					Flair: p.LinkFlair,
				})
			}

			return writeJSON(cmd, map[string]any{
				"ok":        true,
				"subreddit": "r/" + args[0],
				"listing":   listing,
				"count":     len(items),
				"items":     items,
			})
		},
	}
	cmd.Flags().StringVarP(&listing, "listing", "l", "hot", "hot|new|top|rising|controversial")
	cmd.Flags().IntVarP(&limit, "limit", "n", 25, "max posts (max 100)")
	return cmd
}

func newRedditReadCommand() *cobra.Command {
	var postID string
	cmd := &cobra.Command{
		Use:   "read [permalink or post ID]",
		Short: "Read a post and its comments by permalink or ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rc := search.NewRedditClient()
			post, comments, err := rc.PostComments(args[0])
			if err != nil {
				return writeError(cmd, "reddit_error", err.Error(), "")
			}

			type comment struct {
				Author string `json:"author"`
				Body   string `json:"body"`
				Score  int    `json:"score"`
				Date   string `json:"date"`
			}
			var items []comment
			for _, c := range comments {
				items = append(items, comment{
					Author: c.Author, Body: c.Title,
					Score: c.Score,
					Date:  time.Unix(int64(c.Created), 0).Format("2006-01-02"),
				})
			}

			snippet := post.Selftext
			if len(snippet) > 500 {
				snippet = snippet[:500] + "..."
			}

			return writeJSON(cmd, map[string]any{
				"ok": true,
				"post": map[string]any{
					"title": post.Title, "url": "https://www.reddit.com" + post.Permalink,
					"author": post.Author, "score": post.Score,
					"comments_count": post.NumComments, "text": snippet,
					"subreddit": post.Subreddit,
				},
				"comments":       items,
				"comments_count": len(items),
			})
		},
	}
	cmd.Flags().StringVarP(&postID, "id", "i", "", "post ID (t3_xxx)")
	return cmd
}

func newRedditInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info NAME",
		Short: "Subreddit info — subscribers, description, stats",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rc := search.NewRedditClient()
			info, err := rc.SubredditInfo(args[0])
			if err != nil {
				return writeError(cmd, "reddit_error", err.Error(), "")
			}

			desc := info.PublicDesc
			if desc == "" {
				desc = info.Description
			}
			if len(desc) > 500 {
				desc = desc[:500] + "..."
			}

			return writeJSON(cmd, map[string]any{
				"ok": true,
				"subreddit": map[string]any{
					"name":        info.DisplayName,
					"title":       info.Title,
					"url":         info.URL,
					"description": desc,
					"subscribers": info.Subscribers,
					"active":      info.ActiveUserCount,
					"created":     time.Unix(int64(info.Created), 0).Format("2006-01-02"),
					"over18":      info.Over18,
					"language":    info.Lang,
				},
			})
		},
	}
	return cmd
}
