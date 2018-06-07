package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"net/http"
	"encoding/json"
	"io/ioutil"
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

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhookWithCert(config.Domain + bot.Token, "cert.pem"))
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
	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)

	for update := range updates {
		log.Printf("%+v\n", update)
	}

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
