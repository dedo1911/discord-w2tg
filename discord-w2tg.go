package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/leaanthony/clir"
	"github.com/nickname32/discordhook"
)

func main() {
	cli := clir.NewCli("discord-w2tg", "Discord Webhook to Telegram", "v1.0.0")
	port := "8080"
	cli.StringFlag("port", "HTTP listen port", &port)

	showFooter := false
	cli.BoolFlag("footer", "Show footer", &showFooter)

	cli.Action(func() error {
		r := chi.NewRouter()
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		r.Post("/{botID}/{channelID}", func(w http.ResponseWriter, r *http.Request) {
			botID := chi.URLParam(r, "botID")
			channelID, err := strconv.ParseInt(chi.URLParam(r, "channelID"), 10, 64)
			if err != nil {
				log.Println("Error:", err)
				return
			}
			if botID == "" || channelID == 0 {
				log.Println("botID or channelID are empty")
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				return
			}
			var webhook discordhook.WebhookExecuteParams
			err = json.Unmarshal(body, &webhook)
			if err != nil {
				log.Println(err)
				return
			}

			bot, err := tgbot.NewBotAPI(botID)
			if err != nil {
				log.Println(err)
				return
			}
			text := ""
			if webhook.Content != "" {
				text += "<b>" + webhook.Content + "</b>"
			}
			if webhook.Embeds != nil {
				for _, embed := range webhook.Embeds {
					if embed.Title != "" {
						text += "\n<b>" + embed.Title + "</b>"
					}
					if embed.Description != "" {
						text += "\n" + embed.Description
					}
					if embed.URL != "" {
						text += "\n" + embed.URL
					}
					if embed.Fields != nil {
						for _, field := range embed.Fields {
							text += "\n<b>" + field.Name + "</b>: " + strings.ReplaceAll(field.Value, "||", "")
						}
					}
					if showFooter && embed.Footer != nil {
						text += "\n\n<i>" + embed.Footer.Text + "</i>"
					}
				}
			}
			botMsg := tgbot.NewMessage(channelID, text)
			botMsg.ParseMode = "HTML"
			botMsg.DisableWebPagePreview = true
			if _, err := bot.Send(botMsg); err != nil {
				log.Println(err)
				return
			}
		})

		log.Println("Start listening on port:", port)
		return http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	})

	if err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
