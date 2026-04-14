package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/claudex/claudex-cli/internal/auth"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Unbind this machine and revoke session",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := auth.Logout(apiClient, cfg)
			if err != nil {
				return err
			}

			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("\n  %s Logged out\n", green("✓"))
			fmt.Printf("  %s Device unbound (%d/%d devices remaining)\n\n", green("✓"), resp.DevicesUsed, resp.DevicesLimit)

			return nil
		},
	}
}
