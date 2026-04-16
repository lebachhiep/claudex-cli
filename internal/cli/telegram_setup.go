package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"

	"github.com/lebachhiep/claudex-cli/internal/i18n"
	"github.com/lebachhiep/claudex-cli/internal/notification"
)

// runTelegramSetup handles interactive Telegram bot_token + chat_id input
// with back/cancel support and auto-detect chat_id via getUpdates API.
// Returns empty strings if user cancels at any step.
func runTelegramSetup() (botToken, chatID string, err error) {
	// Step 1: Bot Token — empty input = back/cancel
	var token string
	if err = huh.NewInput().
		Title(i18n.T("notify.bot_token")).
		Description(i18n.T("notify.bot_token_desc") + " (empty = " + i18n.T("notify.back") + ")").
		Value(&token).
		Run(); err != nil {
		return "", "", err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", "", nil // cancelled
	}

	// Step 2: Chat ID method
	var method string
	if err = huh.NewSelect[string]().
		Title(i18n.T("notify.chat_method")).
		Options(
			huh.NewOption(i18n.T("notify.chat_auto"), "auto"),
			huh.NewOption(i18n.T("notify.chat_manual"), "manual"),
			huh.NewOption(i18n.T("notify.back"), "back"),
		).
		Value(&method).
		Run(); err != nil {
		return "", "", err
	}

	switch method {
	case "back":
		return "", "", nil
	case "manual":
		id, err := promptManualChatID()
		if err != nil || id == "" {
			return "", "", err
		}
		return token, id, nil
	case "auto":
		id, err := promptAutoDetectChatID(token)
		if err != nil {
			return "", "", err
		}
		if id == "" {
			// Auto-detect cancelled — fallback to manual
			id, err = promptManualChatID()
			if err != nil || id == "" {
				return "", "", err
			}
		}
		return token, id, nil
	}

	return "", "", nil
}

// promptManualChatID asks user to type chat_id directly.
func promptManualChatID() (string, error) {
	var id string
	err := huh.NewInput().
		Title(i18n.T("notify.chat_id")).
		Description(i18n.T("notify.chat_id_desc") + " (empty = " + i18n.T("notify.back") + ")").
		Value(&id).
		Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(id), nil
}

// promptAutoDetectChatID: shows instruction, fetches chats via getUpdates,
// lets user pick. Returns empty string if user backs out.
func promptAutoDetectChatID(botToken string) (string, error) {
	yellow := color.New(color.FgYellow).SprintFunc()

	for {
		// Tell user to message the bot, then wait for Enter
		var confirm bool
		if err := huh.NewConfirm().
			Title(i18n.T("notify.chat_auto_hint")).
			Affirmative("Fetch").
			Negative(i18n.T("notify.back")).
			Value(&confirm).
			Run(); err != nil {
			return "", err
		}
		if !confirm {
			return "", nil
		}

		fmt.Printf("\n  %s\n", i18n.T("notify.chat_fetching"))
		chats, err := notification.FetchRecentChats(botToken)
		if err != nil {
			fmt.Printf("\n  %s %s\n", yellow("!"), fmt.Sprintf(i18n.T("notify.fetch_err"), err))
			// Offer retry or back
			var retry string
			if err := huh.NewSelect[string]().
				Options(
					huh.NewOption(i18n.T("notify.chat_retry"), "retry"),
					huh.NewOption(i18n.T("notify.back"), "back"),
				).
				Value(&retry).
				Run(); err != nil {
				return "", err
			}
			if retry == "back" {
				return "", nil
			}
			continue
		}

		if len(chats) == 0 {
			fmt.Printf("\n  %s %s\n", yellow("!"), i18n.T("notify.chat_empty"))
			var retry string
			if err := huh.NewSelect[string]().
				Options(
					huh.NewOption(i18n.T("notify.chat_retry"), "retry"),
					huh.NewOption(i18n.T("notify.back"), "back"),
				).
				Value(&retry).
				Run(); err != nil {
				return "", err
			}
			if retry == "back" {
				return "", nil
			}
			continue
		}

		// Build select options from chats
		var opts []huh.Option[string]
		for _, c := range chats {
			opts = append(opts, huh.NewOption(notification.FormatChatLabel(c), strconv.FormatInt(c.ID, 10)))
		}
		opts = append(opts, huh.NewOption(i18n.T("notify.back"), "back"))

		var selected string
		if err := huh.NewSelect[string]().
			Title(i18n.T("notify.chat_select")).
			Options(opts...).
			Value(&selected).
			Run(); err != nil {
			return "", err
		}
		if selected == "back" {
			return "", nil
		}
		return selected, nil
	}
}
