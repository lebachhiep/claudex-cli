package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lebachhiep/claudex-cli/internal/auth"
	"github.com/lebachhiep/claudex-cli/internal/i18n"
)

func newLoginCmd() *cobra.Command {
	var licenseKey string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a license key and bind this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if licenseKey == "" {
				return fmt.Errorf(i18n.T("login.key_required"))
			}

			_, resp, err := auth.Login(apiClient, licenseKey, cfg)
			if err != nil {
				return fmt.Errorf("%s", mapAuthError(err))
			}

			green := color.New(color.FgGreen).SprintFunc()

			fmt.Printf("\n  %s %s\n", green("✓"), i18n.T("login.success"))
			fmt.Printf("    %s\n", i18n.T("login.license", color.CyanString(resp.Plan)))
			fmt.Printf("    %s\n", i18n.T("login.machine", resp.DevicesUsed, resp.DevicesLimit))
			fmt.Printf("    %s\n\n", i18n.T("login.lifetime"))

			return nil
		},
	}

	cmd.Flags().StringVar(&licenseKey, "key", "", "License key (required)")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

// mapAuthError converts API errors to user-friendly messages.
func mapAuthError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "invalid_key"):
		return i18n.T("login.err_invalid_key")
	case strings.Contains(msg, "key_inactive"):
		return i18n.T("login.err_inactive")
	case strings.Contains(msg, "device_limit_exceeded"):
		return i18n.T("login.err_device_limit")
	case strings.Contains(msg, "cannot reach"):
		return i18n.T("login.err_network")
	default:
		return msg
	}
}
