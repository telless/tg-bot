package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"net/http"
	"fmt"
)

const config_path string = "config.json"

type baseConfig struct {
	Token  string `json:"token"`
	Domain string `json:"domain"`
}

func main() {
	// make config
	config := parseConfig()

	// init bot
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Fatal(err)
	}

	// enable debug mode
	bot.Debug = true

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	// init webhook for tg api
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(config.Domain + bot.Token))
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	// check for last error
	if info.LastErrorDate != 0 {
		log.Printf("[Telegram callback failed]%s", info.LastErrorMessage)
	}

	// start http server
	go http.ListenAndServe("0.0.0.0:3000", nil)

	// create channel for incoming messages
	updates := bot.ListenForWebhook("/" + bot.Token)

	// listen for OS signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		os.Exit(0)
	}(signals)

	// make closeChannel
	closeChan := make(chan bool)

	// run goroutine for processing inc messages
	go processUpdates(bot, updates, closeChan)

	// end of application
	<-closeChan
}

// process messages
func processUpdates(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, closeChan chan bool) {
	for update := range updates {
		log.Printf("Get message %s from %s\n", update.Message.Text, update.Message.From.String())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg.Text = fmt.Sprintf("Привет %s! Доступные команды /teach, /check, /author", update.Message.From.String())
			case "teach":
				msg.Text = "Тут должен быть текст обучающего урока"
			case "check":
				msg.Text = "Тут должен быть текст вопроса"
			case "author":
				msg.Text = "Arseniy Skurt @skurtars"
			case "muse": // some easter egg here
				msg.Text = "<3"
			default:
				msg.Text = "Попробуй /teach, /check или /author"
			}
		} else {
			msg.Text = "Попробуй /teach, /check или /author"
		}
		bot.Send(msg)
	}

	closeChan <- true
}

// read config
func parseConfig() baseConfig {
	config := baseConfig{}
	content, err := ioutil.ReadFile(config_path)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(content, &config)

	return config
}
