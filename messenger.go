package tgmessenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Messenger struct {
	// if not sending messages to a user, you must include the -100 prefix
	chatID    string
	topicID   int64
	parseMode ParseMode
	botToken  string
}

// NewMessenger returns a new Messenger instance, optionally validating the bot token and chat ID.
// set topicID to -1 if not sending messages to a supergroup.
func NewMessenger(botToken, chatID string, topicID int64, parseMode ParseMode, validate bool) (*Messenger, error) {
	if !validate {
		return &Messenger{
			chatID:    chatID,
			topicID:   topicID,
			parseMode: parseMode,
			botToken:  botToken,
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

	if topicID != -1 {
		body, err := io.ReadAll(chatResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read chat response body: %w", err)
		}

		var chat struct {
			Result struct {
				Type string `json:"type"`
			} `json:"result"`
		}
		if err := json.Unmarshal(body, &chat); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chat response: %w", err)
		}

		if chat.Result.Type != "supergroup" {
			return nil, fmt.Errorf("chat type must be supergroup for topic support, got: %s", chat.Result.Type)
		}
	}

	return &Messenger{
		chatID:    chatID,
		topicID:   topicID,
		parseMode: parseMode,
		botToken:  botToken,
	}, nil
}

func (m Messenger) SendMessage(text string) error {
	payloadData := map[string]any{
		"chat_id":    m.chatID,
		"text":       text,
		"parse_mode": m.parseMode,
	}
	if m.topicID != -1 {
		payloadData["message_thread_id"] = m.topicID
	}

	payload, err := json.Marshal(payloadData)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got non-200 status code: %s; body: %s", resp.Status, string(body))
	}

	return nil
}
