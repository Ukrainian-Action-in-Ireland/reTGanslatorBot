package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	Chats []Chat `json:"chats"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("BOT_TOKEN has to be specified")
	}

	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config.json: %v", err)
	}

	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatalf("Failed to read config.json: %v", err)
	}

	var config Config
	if err = json.Unmarshal(configBytes, &config); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Bot API failed to initialize: %v", err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		found := false
		for _, chat := range config.Chats {
			if update.Message.Chat.ID == chat.ID {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		for _, chat := range config.Chats {
			if !strings.Contains(strings.ToLower(update.Message.Text), "@"+strings.ToLower(chat.Name)) {
				continue
			}

			msg := tgbotapi.NewMessage(chat.ID, update.Message.Text)
			bot.Send(msg)
		}
	}
}
