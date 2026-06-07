package cli

import (
	"fmt"
	"os"

	"github.com/izzzzzi/agent-asearch/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration — API keys, defaults",
	}
	cmd.AddCommand(newConfigShowCommand())
	cmd.AddCommand(newConfigSetCommand())
	cmd.AddCommand(newConfigGetCommand())
	cmd.AddCommand(newConfigCompletionCommand())
	return cmd
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration (keys masked)",
		Args:  noPositionalArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return writeError(cmd, "config_error", err.Error(), "")
			}

			masked := make(map[string]string)
			for k, v := range cfg.APIKeys {
				if len(v) > 8 {
					masked[k] = v[:4] + "…" + v[len(v)-4:]
				} else if v != "" {
					masked[k] = "***"
				}
			}

			return writeJSON(cmd, map[string]any{
				"ok":       true,
				"config":   masked,
				"defaults": cfg.Defaults,
			})
		},
	}
}

func newConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set an API key or default",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			if key == "default.limit" {
				cfg, err := config.Load()
				if err != nil {
					return writeError(cmd, "config_error", err.Error(), "")
				}
				var n int
				if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
					return writeError(cmd, "invalid_args", "limit must be a number", "")
				}
				cfg.Defaults.Limit = n
				if err := config.Save(cfg); err != nil {
					return writeError(cmd, "config_error", err.Error(), "")
				}
				return writeJSON(cmd, map[string]any{"ok": true, "key": key, "value": n})
			}

			if err := config.SetKey(key, value); err != nil {
				return writeError(cmd, "config_error", err.Error(), "")
			}

			masked := value
			if len(masked) > 8 {
				masked = masked[:4] + "…" + masked[len(masked)-4:]
			} else if masked != "" {
				masked = "***"
			}

			return writeJSON(cmd, map[string]any{
				"ok": true, "key": key, "value": masked,
				"hint": fmt.Sprintf("Key saved to %s", config.Path()),
			})
		},
	}
}

func newConfigGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			// Check env first
			envKey := configKeyToEnv(key)
			if envKey != "" {
				if v := os.Getenv(envKey); v != "" {
					masked := v
					if len(masked) > 8 {
						masked = masked[:4] + "…" + masked[len(masked)-4:]
					}
					return writeJSON(cmd, map[string]any{
						"ok": true, "key": key, "value": masked, "source": "env",
					})
				}
			}

			cfg, err := config.Load()
			if err != nil {
				return writeError(cmd, "config_error", err.Error(), "")
			}
			if v, ok := cfg.APIKeys[key]; ok {
				masked := v
				if len(masked) > 8 {
					masked = masked[:4] + "…" + masked[len(masked)-4:]
				}
				return writeJSON(cmd, map[string]any{
					"ok": true, "key": key, "value": masked, "source": "config",
				})
			}

			return writeJSON(cmd, map[string]any{
				"ok": false, "key": key, "message": "not found",
				"hint": fmt.Sprintf("Set with: asearch config set %s <value>", key),
			})
		},
	}
}

func configKeyToEnv(key string) string {
	m := map[string]string{
		"tavily": "TAVILY_API_KEY", "exa": "EXA_API_KEY",
		"brave": "BRAVE_API_KEY", "serper": "SERPER_API_KEY",
		"serpapi": "SERPAPI_API_KEY", "perplexity": "PERPLEXITY_API_KEY",
		"you": "YOU_API_KEY", "firecrawl": "FIRECRAWL_API_KEY",
		"parallel": "PARALLEL_API_KEY", "searxng": "ASEARCH_SEARXNG_URL",
	}
	return m[key]
}

func newConfigCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			var err error
			switch shell {
			case "bash":
				err = cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				err = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				err = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			default:
				return writeError(cmd, "invalid_args", "unsupported shell: "+shell, "bash|zsh|fish")
			}
			if err != nil {
				return writeError(cmd, "completion_error", err.Error(), "")
			}
			return nil
		},
	}
}
