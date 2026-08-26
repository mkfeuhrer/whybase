package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagRoot string
	stderr   = os.Stderr
)

func NewRoot() *cobra.Command {
	c := &cobra.Command{
		Use:           "whybase",
		Short:         "Every why in your codebase, one query away",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	c.PersistentFlags().StringVar(&flagRoot, "dir", ".", "repository root")
	c.AddCommand(newCmd(), listCmd(), supersedeCmd())
	return c
}

func Execute() error {
	if err := NewRoot().Execute(); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return err
	}
	return nil
}
