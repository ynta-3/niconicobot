package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	traq "github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	payload "github.com/traPtitech/traq-ws-bot/payload"
)

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

func main() {
	// Load environment variables from .env when present.
	_ = godotenv.Load()

	accessToken := os.Getenv("TRAQ_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("TRAQ_ACCESS_TOKEN is not set")
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
			case "\\ping":
				content = "pong " + args
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
