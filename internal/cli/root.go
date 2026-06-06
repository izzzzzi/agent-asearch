package cli

import (
	"errors"

	"github.com/izzzzzi/agent-asearch/internal/agenthelp"
	"github.com/izzzzzi/agent-asearch/internal/response"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "asearch",
		Short:         "Search workflow CLI for LLM agents",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return writeInvalidArgs(cmd, "unknown command "+args[0], "run asearch --help")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeInvalidArgs(cmd, "command required", "run asearch --help")
		},
	}
	cmd.PersistentFlags().Bool("json", true, "emit JSON output (deprecated; JSON is always emitted)")
	_ = cmd.PersistentFlags().MarkHidden("json")
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return writeInvalidArgs(cmd, err.Error(), "run asearch --help")
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		writeAgentHelp(cmd.Root())
	})

	cmd.AddCommand(
		newPromptCommand(),
		newOpenCommand(),
		newResultsCommand(),
		newSessionCommand(),
		newDoctorCommand(),
		newVersionCommand(),
	)
	setCommandHelp(cmd)
	return cmd
}

func Execute() error {
	return NewRootCommand().Execute()
}

func noPositionalArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return writeInvalidArgs(cmd, "unexpected positional arguments", "")
	}
	return nil
}

func writeInvalidArgs(cmd *cobra.Command, message, hint string) error {
	body, marshalErr := response.MarshalError("invalid_args", message, hint)
	if marshalErr != nil {
		return marshalErr
	}
	_, _ = cmd.ErrOrStderr().Write(body)
	return errors.New(message)
}

func writeJSON(cmd *cobra.Command, v any) error {
	body, err := response.Marshal(v)
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(body)
	return err
}

func writeCommandHelp(cmd *cobra.Command, args []string) {
	if _, err := cmd.OutOrStdout().Write([]byte(cmd.UsageString())); err != nil {
		_, _ = cmd.ErrOrStderr().Write([]byte("failed to write help: " + err.Error() + "\n"))
	}
}

func setCommandHelp(cmd *cobra.Command) {
	for _, child := range cmd.Commands() {
		child.SetHelpFunc(writeCommandHelp)
		setCommandHelp(child)
	}
}

func writeError(cmd *cobra.Command, code, message, hint string) error {
	body, marshalErr := response.MarshalError(code, message, hint)
	if marshalErr != nil {
		return marshalErr
	}
	_, _ = cmd.ErrOrStderr().Write(body)
	return errors.New(message)
}

func writeAgentHelp(root *cobra.Command) {
	payload := agenthelp.RootHelp()
	if err := writeJSON(root, payload); err != nil {
		_, _ = root.ErrOrStderr().Write([]byte(`{"ok":false,"code":"help_error","message":"failed to emit help"}` + "\n"))
	}
}
