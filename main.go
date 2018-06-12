package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"net/http"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

const configPath = "config.json"

func main() {
	// make config
	config := parseConfig()

	// init bot
	bot, err := tgbotapi.NewBotAPI(config.Token)
	failOnError(err)

	os.Stdout.WriteString(fmt.Sprintf("Authorized on account %s\n", bot.Self.UserName))

	// init webhook for tg api
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(config.Domain + bot.Token))
	failOnError(err)

	info, err := bot.GetWebhookInfo()
	failOnError(err)

	// check for last error
	if info.LastErrorDate != 0 {
		os.Stdout.WriteString(fmt.Sprintf("[Telegram callback failed]%s", info.LastErrorMessage))
	}

	// start http server
	go http.ListenAndServe("0.0.0.0:3000", nil)

	// create channel for incoming messages
	updates := bot.ListenForWebhook("/" + bot.Token)

	// make application instance
	closeChan := make(chan bool)
	users := initUsers()
	lessons := initLessons()
	app := Application{bot, updates, closeChan, config, users, lessons}

	// run goroutine for processing inc messages
	go app.run()

	// end of application
	<-closeChan
}

// read config
func parseConfig() Config {
	config := Config{}
	content, err := ioutil.ReadFile(configPath)
	failOnError(err)
	json.Unmarshal(content, &config)

	return config
}

func failOnError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func logError(err error) {
	if err != nil {
		log.Print(err)
	}
}
