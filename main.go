package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

const config_path string = "config.json"

type baseConfig struct {
	Token  string `json:"token"`
	Domain string `json:"domain"`
}

func main() {
	config := parseConfig()
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(config.Domain + bot.Token))
	if err != nil {
		log.Fatal(err)
	}
	info, err := bot.GetWebhookInfo()

	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("[Telegram callback failed]%s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		os.Exit(0)
	}(signals)

	go http.ListenAndServe("0.0.0.0:3000", nil)
	closeChan := make(chan bool)
	go processUpdates(bot, updates, closeChan)

	<-closeChan

}

func processUpdates(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, closeChan chan bool) {
	for update := range updates {
		user := update.Message.From.FirstName + " " + update.Message.From.LastName + "(aka " + update.Message.From.UserName + ")"
		log.Printf("Get message %s from %s\n", update.Message.Text, user)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello "+user+", thx for "+update.Message.Text)

		bot.Send(msg)
	}

	closeChan <- true
}

func parseConfig() baseConfig {
	config := baseConfig{}
	content, err := ioutil.ReadFile(config_path)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(content, &config)

	return config
}
