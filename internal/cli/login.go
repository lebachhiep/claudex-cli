package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/claudex/claudex-cli/internal/auth"
)

func newLoginCmd() *cobra.Command {
	var licenseKey string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a license key and bind this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if licenseKey == "" {
				return fmt.Errorf("license key is required. Use --key=YOUR_KEY")
			}

			_, resp, err := auth.Login(apiClient, licenseKey, cfg)
			if err != nil {
				return fmt.Errorf("%s", mapAuthError(err))
			}

			green := color.New(color.FgGreen).SprintFunc()

			fmt.Printf("\n  %s Authenticated successfully\n", green("✓"))
			fmt.Printf("    License:  %s\n", color.CyanString(resp.Plan))
			fmt.Printf("    Machine:  bound (%d/%d devices)\n", resp.DevicesUsed, resp.DevicesLimit)
			fmt.Printf("    License:  Lifetime\n\n")

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
		return "Invalid license key"
	case strings.Contains(msg, "key_inactive"):
		return "License has been deactivated"
	case strings.Contains(msg, "device_limit_exceeded"):
		return "Device limit reached. Run `claudex logout` on another device to free a slot"
	case strings.Contains(msg, "cannot reach"):
		return "Cannot reach API server. Check your internet connection"
	default:
		return msg
	}
}
