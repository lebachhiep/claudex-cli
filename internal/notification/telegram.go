package notification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"
)

// TelegramChat represents a chat from Telegram getUpdates response.
type TelegramChat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"` // private | group | supergroup | channel
	Title     string `json:"title,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// telegramUpdatesResponse is the top-level JSON shape from getUpdates.
type telegramUpdatesResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
	Result      []struct {
		Message *struct {
			Chat TelegramChat `json:"chat"`
		} `json:"message,omitempty"`
		ChannelPost *struct {
			Chat TelegramChat `json:"chat"`
		} `json:"channel_post,omitempty"`
	} `json:"result"`
}

// FetchRecentChats calls Telegram's getUpdates API and returns deduplicated chats.
// Returns error if token invalid, network fails, or API returns ok=false.
func FetchRecentChats(botToken string) ([]TelegramChat, error) {
	if botToken == "" {
		return nil, fmt.Errorf("bot token is empty")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", url.PathEscape(botToken))
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	var parsed telegramUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if !parsed.OK {
		if parsed.Description != "" {
			return nil, fmt.Errorf("telegram: %s", parsed.Description)
		}
		return nil, fmt.Errorf("telegram API returned not ok")
	}

	// Deduplicate by chat ID
	seen := make(map[int64]TelegramChat)
	for _, update := range parsed.Result {
		var chat *TelegramChat
		if update.Message != nil {
			chat = &update.Message.Chat
		} else if update.ChannelPost != nil {
			chat = &update.ChannelPost.Chat
		}
		if chat != nil && chat.ID != 0 {
			seen[chat.ID] = *chat
		}
	}

	chats := make([]TelegramChat, 0, len(seen))
	for _, c := range seen {
		chats = append(chats, c)
	}

	// Sort: private first, then by name
	sort.Slice(chats, func(i, j int) bool {
		if chats[i].Type != chats[j].Type {
			return chats[i].Type == "private"
		}
		return chatDisplayName(chats[i]) < chatDisplayName(chats[j])
	})

	return chats, nil
}

// FormatChatLabel returns a human-friendly one-line description of a chat.
func FormatChatLabel(chat TelegramChat) string {
	name := chatDisplayName(chat)
	if chat.Username != "" {
		return fmt.Sprintf("%s (@%s) [%s, id: %d]", name, chat.Username, chat.Type, chat.ID)
	}
	return fmt.Sprintf("%s [%s, id: %d]", name, chat.Type, chat.ID)
}

// chatDisplayName returns the best name to display for a chat.
func chatDisplayName(chat TelegramChat) string {
	if chat.Title != "" {
		return chat.Title
	}
	name := chat.FirstName
	if chat.LastName != "" {
		if name != "" {
			name += " " + chat.LastName
		} else {
			name = chat.LastName
		}
	}
	if name == "" {
		name = fmt.Sprintf("Chat %d", chat.ID)
	}
	return name
}
