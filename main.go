package main

import (        
        "context"
        "log"

        traq "github.com/traPtitech/go-traq"
        traqwsbot "github.com/traPtitech/traq-ws-bot"
        payload "github.com/traPtitech/traq-ws-bot/payload"
)

type stampList sturuct {
        userId string `json:"userId"`
        stampId string `json:"stampId"`
        count int `json:"count"`
        createdAt string `json:"createdAt"`
        updatedAt string `json:"updatedAt"`
}

// traQのAPIからスタンプの情報を取得する関数
func getStampListFromApi(url string) ([]message, error) {
        resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
}

func main() {
        bot, err := traqwsbot.NewBot(&traqwsbot.Options{
                AccessToken: "CYclBElpFyELEvoeTehdHWkAKioLXiATzSSs",
        })
        if err != nil {
                panic(err)
        }

        bot.OnMessageCreated(func(p *payload.MessageCreated) {
                log.Println("Received MESSAGE_CREATED event: " + p.Message.Text)
                _, _, err := bot.API().
                        MessageApi.
                        PostMessage(context.Background(), p.Message.ChannelID).
                        PostMessageRequest(traq.PostMessageRequest{
                                Content: "oisu-",
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