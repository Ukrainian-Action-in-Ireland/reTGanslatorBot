package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	Chats []Chat `json:"chats"`
}

type Chat struct {
	ID      int64    `json:"id"`
	Aliases []string `json:"aliases"`
}

type BotHandlers struct {
	bot    *tgbotapi.BotAPI
	config Config
}

func hasChatTag(chatName, text string) bool {
	return strings.Contains(strings.ToLower(text), "*"+strings.ToLower(chatName))
}

func (config Config) AllAliases() []string {
	aliases := make(map[string]bool)
	for _, chat := range config.Chats {
		for _, alias := range chat.Aliases {
			aliases[alias] = true
		}
	}
	var aliasesList []string
	for alias := range aliases {
		aliasesList = append(aliasesList, alias)
	}
	sort.Slice(aliasesList, func(i, j int) bool {
		return strings.ToLower(aliasesList[i]) < strings.ToLower(aliasesList[j])
	})
	return aliasesList
}

func (bh BotHandlers) inlineQuery(update tgbotapi.Update) {
	query := *update.InlineQuery
	aliases := bh.config.AllAliases()
	for i, alias := range aliases {
		aliases[i] = "*" + alias
	}
	matched := aliases
	words := strings.Fields(query.Query)
	withoutLastWord := query.Query
	if len(words) > 0 {
		matched = nil
		lastWord := words[len(words)-1]
		withoutLastWord = strings.TrimRightFunc(query.Query, unicode.IsSpace)[0 : len(query.Query)-len(lastWord)]
		for _, alias := range aliases {
			if strings.Contains(strings.ToLower(alias), strings.ToLower(lastWord)) {
				matched = append(matched, alias)
			}
		}
	}
	var results []interface{}
	for _, alias := range matched {
		fullSuggestion := withoutLastWord + alias
		results = append(results, tgbotapi.NewInlineQueryResultArticle(fullSuggestion, alias, fullSuggestion))
	}
	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       results,
	}
	bh.bot.AnswerInlineQuery(inlineConfig)
}

func (bh BotHandlers) message(update tgbotapi.Update) {
	log.Printf("[%s] text: %s, caption: %s", update.Message.From.UserName, update.Message.Text, update.Message.Caption)
	if update.Message.Entities != nil {
		for _, entity := range *update.Message.Entities {
			username := update.Message.Text[entity.Offset : entity.Offset+entity.Length]
			if entity.Type == "mention" && username == "@reTGanslatorBot" {
				aliases := bh.config.AllAliases()
				for i, alias := range aliases {
					aliases[i] = "*" + strings.ToLower(alias)
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Tags: "+strings.Join(aliases, " "))
				msg.BaseChat.ReplyToMessageID = update.Message.MessageID
				bh.bot.Send(msg)
			}
		}
	}

	found := false
	for _, chat := range bh.config.Chats {
		if update.Message.Chat.ID == chat.ID {
			found = true
			break
		}
	}
	if !found {
		return
	}

	for _, chat := range bh.config.Chats {
		hasTags := false
		for _, alias := range chat.Aliases {
			if hasChatTag(alias, update.Message.Text) || hasChatTag(alias, update.Message.Caption) {
				hasTags = true
				break
			}
		}
		if !hasTags {
			continue
		}

		{
			msg := tgbotapi.NewMessage(chat.ID, "Пересилаю повідомлення з чату "+update.Message.Chat.Title)
			bh.bot.Send(msg)
		}
		if update.Message.ReplyToMessage != nil {
			msg := tgbotapi.NewForward(chat.ID, update.Message.Chat.ID, update.Message.ReplyToMessage.MessageID)
			bh.bot.Send(msg)
		}
		{
			msg := tgbotapi.NewForward(chat.ID, update.Message.Chat.ID, update.Message.MessageID)
			bh.bot.Send(msg)
		}
	}
}

func (bh BotHandlers) command(update tgbotapi.Update) {
	aliases := bh.config.AllAliases()
	for i := range aliases {
		aliases[i] = "*" + strings.ToLower(aliases[i])
	}
	aliasesStr := strings.Join(aliases, " ")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(`
Для того, щоб поширити ваше повідомлення по декільком чатам одночасно, y тексті повідомлення напишіть будь-який з наступних тегів з зірочкою, і бот перешле повідомлення до всіх вказаних чатів:
%s

Для того, щоб перекинути вже відправлене повідомлення, відповідайте на це повідомлення і у тексті відповіді вказуйте теги. Бот знайде теги і перекине обидва повідомлення: те, що без тега і на яке відповіли, та вашy відповідь з тегом.

Для того, щоб побачити доступні теги, використовуйте рядок команди бота. Для цього почніть писати повідомлення з "@reTGanslator ", і бот запропонує вам список тегів. Після додавання тексту і всіх тегів натисніть на один з запропонованих тегів для того, щоб відправити повідомлення.
Також можна тегнути бота у будь-якому повідомленні, і бот пришле список усіх тегів.

Для того, щоб додати бота та теги до вашого чата, звертайтесь до адмінів вашого чату.
	`, aliasesStr))
	bh.bot.Send(msg)
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

	botHandlers := BotHandlers{bot, config}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		switch {
		case update.InlineQuery != nil:
			botHandlers.inlineQuery(update)
		case update.Message != nil && update.Message.IsCommand():
			botHandlers.command(update)
		case update.Message != nil:
			botHandlers.message(update)
		default:
			log.Printf("Unknown type of message")
		}
	}
}
