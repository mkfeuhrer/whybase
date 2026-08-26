package cli

import (
	"github.com/mkfeuhrer/whybase/internal/mcpserver"
	"github.com/spf13/cobra"
)

func mcpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Serve decision records to AI agents over MCP (stdio)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ix, _, ferrs, err := loadIndex(flagRoot)
			if err != nil {
				return err
			}
			for _, fe := range ferrs {
				logWarn(fe.Error())
			}
			return mcpserver.Run(cmd.Context(), ix)
		},
	}
}
