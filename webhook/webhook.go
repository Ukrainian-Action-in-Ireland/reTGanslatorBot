package webhook

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/DzyubSpirit/reTGanslatorBot/bot"
	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	botHandler      *bot.Handler
	botWebhookToken string
)

func init() {
	functions.HTTP("UpdateHandler", updateHandler)
}

func init() {
	botWebhookToken = os.Getenv("WEBHOOK_TOKEN")

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("BOT_TOKEN has to be specified")
	}

	path := "serverless_function_source_code/config.json"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Fall back to the current working directory if that file doesn't exist.
		path = "config.json"
	}

	configFile, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config.json: %v", err)
	}

	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatalf("Failed to read config.json: %v", err)
	}

	var config bot.Config
	if err = json.Unmarshal(configBytes, &config); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}

	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Bot API failed to initialize: %v", err)
	}

	botHandler = bot.NewHandler(config, tgBot)

	tgBot.Debug = true

	log.Printf("Authorized on account %s", tgBot.Self.UserName)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Unknown HTTP method: %s", r.Method)
		httpErr(w, http.StatusMethodNotAllowed)
		return
	}

	if path := strings.Trim(r.URL.Path, "/"); path != botWebhookToken {
		log.Printf("Wrong path %s", path)
		httpErr(w, http.StatusNotFound)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read the incoming payload: %v", err)
		httpErr(w, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var update tgbotapi.Update
	if err := json.Unmarshal(data, &update); err != nil {
		log.Printf("Failed to unmarshal incoming update: %v", err)
		httpErr(w, http.StatusBadRequest)
		return
	}

	if err := botHandler.HandleUpdate(update); err != nil {
		log.Printf("Handle incoming update: %v", err)
		httpErr(w, http.StatusInternalServerError)
	}
}

func httpErr(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
