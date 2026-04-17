package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	traq "github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	payload "github.com/traPtitech/traq-ws-bot/payload"
)

const geminiModel = "gemini-2.5-flash"

func parse_commands(text string) (command string, args string, ok bool) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "", "", false
	}

	// When the message starts with a parsed mention token, ignore it.
	if strings.HasPrefix(fields[0], "!{") {
		fields = fields[1:]
	}
	if len(fields) == 0 {
		return "", "", false
	}

	if !strings.HasPrefix(fields[0], "\\") {
		return "", "", false
	}

	command = fields[0]
	args = strings.TrimSpace(strings.Join(fields[1:], " "))
	return command, args, true
}

type geminiGenerateRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func askGemini(ctx context.Context, apiKey string, prompt string) (string, error) {
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		geminiModel,
		url.QueryEscape(apiKey),
	)

	reqBody := geminiGenerateRequest{}
	reqBody.Contents = append(reqBody.Contents, struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	}{
		Parts: []struct {
			Text string `json:"text"`
		}{
			{Text: prompt},
		},
	})

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed geminiGenerateResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("gemini api error: %s", parsed.Error.Message)
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini api error: empty response")
	}

	text := strings.TrimSpace(parsed.Candidates[0].Content.Parts[0].Text)
	if text == "" {
		return "", fmt.Errorf("gemini api error: empty text")
	}

	return text, nil
}

func main() {
	// Load environment variables from .env when present.
	_ = godotenv.Load()

	accessToken := os.Getenv("TRAQ_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("TRAQ_ACCESS_TOKEN is not set")
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}

	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: accessToken,
	})
	if err != nil {
		panic(err)
	}

	bot.OnMessageCreated(func(p *payload.MessageCreated) {
		log.Println("Received MESSAGE_CREATED event: " + p.Message.Text)

		content := "oisu-"
		if cmd, args, ok := parse_commands(p.Message.Text); ok {
			switch cmd {
			case "\\ask":
				if strings.TrimSpace(args) == "" {
					content = "usage: \\ask <prompt>"
					break
				}
				reply, err := askGemini(context.Background(), geminiAPIKey, args)
				if err != nil {
					log.Println(err)
					content = "gemini error"
					break
				}
				content = reply
			case "\\repeat":
				content = args
			case "\\compliance":
			default:
				content = "unknown command: " + cmd
			}
		}

		_, _, err := bot.API().
			MessageAPI.
			PostMessage(context.Background(), p.Message.ChannelID).
			PostMessageRequest(traq.PostMessageRequest{
				Content: content,
			}).
			Execute()
		if err != nil {
			log.Println(err)
		}
	})

	if err := bot.Start(); err != nil {
		panic(err)
	}
}
