package tgmessenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Messenger struct {
	// if not sending messages to a user, you must include the -100 prefix
	ChatID   string
	botToken string
}

// NewMessenger returns a new Messenger instance, optionally validating the bot token and chat ID
func NewMessenger(botToken, chatID string, validate bool) (*Messenger, error) {
	if !validate {
		return &Messenger{
			ChatID:   chatID,
			botToken: botToken,
		}, nil
	}

	// validate token
	tokenResp, err := http.Get(fmt.Sprintf(getMeURL, botToken))
	if err != nil {
		return nil, err
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got non-200 status code when authorizing: %s", tokenResp.Status)
	}

	// validate chatid
	chatResp, err := http.Get(fmt.Sprintf(getChatURL, botToken, chatID))
	if err != nil {
		return nil, err
	}
	defer chatResp.Body.Close()

	if chatResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got non-200 status code when validating chat ID: %s", chatResp.Status)
	}

	return &Messenger{
		ChatID:   chatID,
		botToken: botToken,
	}, nil
}

func (m Messenger) SendMessage(text string) error {
	payload, err := json.Marshal(map[string]string{
		"chat_id": m.ChatID,
		"text":    text,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(sendMessageURL, m.botToken),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got non-200 status code: %s", resp.Status)
	}

	return nil
}
