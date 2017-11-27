package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

const (
	telegramAPIBaseURL = "https://api.telegram.org"

	rGetMe       = "getMe"
	rSendMessage = "sendMessage"
)

var (
	telegramAPIBotToken = os.Getenv("TELEGRAM_API_BOT_TOKEN")
)

type (
	resGetMe struct {
		ID        int
		IsBot     bool   `json:"is_bot"`
		FirstName string `json:"first_name"`
		Username  string
	}

	reqSendMessage struct {
		ChatID           int    `json:"chat_id"`
		Text             string `json:"text"`
		ReplyToMessageID int    `json:"reply_to_message_id"`
		ParseMode        string `json:"parse_mode"`
	}

	resSendMessage mMessage

	mUpdate struct {
		ID      int `json:"update_id"`
		Message mMessage
	}

	mMessage struct {
		ID   int `json:"message_id"`
		From mUser
		Chat mChat
		Date unixTime
		Text string
	}

	mUser struct {
		ID           int
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Username     string
		LanguageCode string `json:"language_code"`
	}

	mChat struct {
		ID        int
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string
		Type      string
	}
)

func main() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var update mUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var (
			req = reqSendMessage{
				ChatID:           update.Message.Chat.ID,
				Text:             "*You just said:* " + update.Message.Text,
				ReplyToMessageID: update.Message.ID,
				ParseMode:        "Markdown",
			}
			res resSendMessage
		)

		if err := request(rSendMessage, &res, req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Print("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func request(method string, data interface{}, params ...interface{}) error {
	var body []byte
	if len(params) > 0 {
		b, err := json.Marshal(params[0])
		if err != nil {
			return err
		}
		body = b
	}

	request, err := http.NewRequest(http.MethodPost, apiURL(method), bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	var out struct {
		Ok          bool
		ErrorCode   int `json:"error_code"`
		Description string
		Result      interface{}
	}

	defer response.Body.Close()
	if err = json.NewDecoder(response.Body).Decode(&out); err != nil {
		return err
	}

	if !out.Ok {
		return fmt.Errorf("got: %s [%d]", out.Description, out.ErrorCode)
	}

	if data == nil {
		return nil
	}

	b, err := json.Marshal(out.Result)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, data)
}

func apiURL(method string) string {
	return fmt.Sprintf("%s/bot%s/%s", telegramAPIBaseURL, telegramAPIBotToken, method)
}

type unixTime int

func (t unixTime) Time() time.Time {
	return time.Unix(int64(t), 0)
}

func (t unixTime) String() string {
	return t.Time().Format(time.RFC3339)
}
