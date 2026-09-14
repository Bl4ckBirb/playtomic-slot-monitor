package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// telegram sends messages to a chat via the Bot API. A zero value (empty
// token/chat) is a no-op that only logs, so a run without Telegram secrets
// still works.
type telegram struct {
	token  string
	chatID string
	http   *http.Client
}

func newTelegram(token, chatID string) *telegram {
	return &telegram{
		token:  token,
		chatID: chatID,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

// configured reports whether both credentials are present.
func (t *telegram) configured() bool {
	return t.token != "" && t.chatID != ""
}

// send posts one message. It returns an error only on a real send failure.
func (t *telegram) send(ctx context.Context, text string) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	form := url.Values{}
	form.Set("chat_id", t.chatID)
	form.Set("text", text)
	form.Set("disable_web_page_preview", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.http.Do(req)
	if err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<10))
		return fmt.Errorf("telegram API %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

// formatMessage renders the new slots grouped by club and day.
func formatMessage(slots []matchedSlot) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("🎾 %d neue Padel-Slots\n", len(slots)))

	var lastClub, lastDay string
	for _, s := range slots {
		club := s.club.Name
		if club == "" {
			club = s.club.TenantID
		}
		day := s.started.Format("Mon 2006-01-02")

		if club != lastClub {
			b.WriteString("\n")
			b.WriteString(club)
			if s.club.URL != "" {
				b.WriteString(" — ")
				b.WriteString(s.club.URL)
			}
			b.WriteString("\n")
			lastClub, lastDay = club, ""
		}
		if day != lastDay {
			b.WriteString("  " + day + "\n")
			lastDay = day
		}

		kind := "outdoor"
		if s.indoor {
			kind = "indoor"
		}
		end := s.started.Add(time.Duration(s.duration) * time.Minute)
		price := s.price
		if price == "" {
			price = "?"
		}
		b.WriteString(fmt.Sprintf("    %s–%s | %s (%s) | %d min | %s\n",
			s.started.Format("15:04"), end.Format("15:04"), s.courtName, kind, s.duration, price))
	}
	return b.String()
}
