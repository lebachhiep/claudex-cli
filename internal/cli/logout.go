package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/i18n"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Unbind this machine and revoke session",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, localOnly, err := auth.Logout(apiClient, cfg)
			if err != nil {
				return err
			}

			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("logout.success"))
			if localOnly {
				fmt.Printf("  %s %s\n\n", green("✓"), i18n.T("logout.local_only"))
			} else {
				fmt.Printf("  %s %s\n\n", green("✓"), i18n.T("logout.unbound", resp.DevicesUsed, resp.DevicesLimit))
			}

			return nil
		},
	}
}
