package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Button struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

func (user User) SendMessage(admin bool, text string, button ...[]Button) {
	text = strings.Replace(text, "_", "\\_", -1)

	message := map[string]interface{}{
		"chat_id":    user.UserID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	if len(button) > 0 {
		message["reply_markup"] = map[string]interface{}{
			"inline_keyboard": button,
		}
	}

	user.sendRequest(admin, "/sendMessage", message)
}

func (user User) SendPhoto(admin bool, url string, text ...string) {
	photo := map[string]interface{}{
		"chat_id": user.UserID,
		"photo":   url,
	}

	if len(text) > 0 {
		photo["caption"] = text[0]
		photo["parse_mode"] = "Markdown"
	}

	user.sendRequest(admin, "/sendPhoto", photo)
}

func (user User) sendRequest(admin bool, endpoint string, payload map[string]interface{}) {
	var apiURL string
	var token string

	if admin {
		token = os.Getenv("TELEGRAM_BOT_TOKEN_ADMIN")
	} else {
		token = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	apiURL = fmt.Sprintf("https://api.telegram.org/bot%s%s", token, endpoint)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON payload: %s", err.Error())
		return
	}

	log.Printf("REQUEST TELEGRAM: url : %s , body: %s", apiURL, jsonPayload)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error sending request: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API request failed with status: %s", resp.Status)
		return
	}

	log.Printf("SUCCESS REQUEST TELEGRAM")
}
