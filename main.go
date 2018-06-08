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
	"time"
	"os/exec"
)

const configPath = "config.json"
const cleverImgPath = "img/clever.png"
const doneImgPath = "img/done.png"
const loadingImgPath = "img/loading.png"

type baseConfig struct {
	Token    string `json:"token"`
	Domain   string `json:"domain"`
	RootPass string `json:"root_pass"`
}

type pictures struct {
	clever  []byte
	done    []byte
	loading []byte
}

type user struct {
	id             int
	username       string
	fullName       string
	lastLessonId   string
	lastVisit      time.Time
	hasAdminRights bool
	authorized     bool
}

func (u *user) applyUpdate(update tgbotapi.Update) {
	u.lastVisit = time.Now()
	u.username = update.Message.From.String()
	u.fullName = fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
}

var (
	users = make(map[int]user)
)

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

	os.Stdout.WriteString(fmt.Sprintf("Authorized on account %s\n", bot.Self.UserName))

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
		os.Stdout.WriteString(fmt.Sprintf("[Telegram callback failed]%s", info.LastErrorMessage))
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
		die(sig)
	}(signals)

	// make closeChannel
	closeChan := make(chan bool)

	// prepare images
	pictures := initPictures()

	// run goroutine for processing inc messages
	go processUpdates(bot, updates, closeChan, pictures, config)

	// end of application
	<-closeChan
}

// process messages
func processUpdates(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, closeChan chan bool, pictures pictures, config baseConfig) {
	for update := range updates {
		user := processUser(update)
		os.Stdout.WriteString(fmt.Sprintf("Get message %s from %s\n", update.Message.Text, update.Message.From.String()))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg.Text = fmt.Sprintf("Привет %s! Доступные команды /teach, /check, /author", update.Message.From.String())
			case "teach":
				bot.Send(tgbotapi.NewPhotoUpload(update.Message.Chat.ID, tgbotapi.FileBytes{Name: "clever.png", Bytes: pictures.clever}))
				msg.Text = "Тут должен быть текст обучающего урока"
			case "check":
				bot.Send(tgbotapi.NewPhotoUpload(update.Message.Chat.ID, tgbotapi.FileBytes{Name: "done.png", Bytes: pictures.done}))
				msg.Text = "Тут должен быть текст вопроса"
			case "author":
				//noinspection SpellCheckingInspection
				msg.Text = "Arseniy Skurt @skurtars"
			case "muse": // some easter egg here
				msg.Text = "<3"
			case "auth": // some easter egg here
				if update.Message.CommandArguments() == config.RootPass {
					user.hasAdminRights = true
					users[user.id] = user
					msg.Text = fmt.Sprintf("Hello %s (%s), you are admin now!", user.fullName, user.username)
				} else {
					msg.Text = "Nice attempt retard"
				}
			case "rebuild":
				if user.hasAdminRights && update.Message.CommandArguments() != "" {
					msg.Text = "Trying to rebuild"
					rebuild(update.Message.CommandArguments(), bot, update)
				} else {
					msg.Text = fmt.Sprintf("%+v attempt to rebuild with branch\tag %s", user, update.Message.CommandArguments())
				}
			case "whoami":
				msg.Text = fmt.Sprintf("Hello %s (%s), here is your data %+v", user.fullName, user.username, user)
			default:
				bot.Send(tgbotapi.NewPhotoUpload(update.Message.Chat.ID, tgbotapi.FileBytes{Name: "loading.png", Bytes: pictures.loading}))
				msg.Text = "Попробуй /teach, /check или /author"
			}
		} else {
			bot.Send(tgbotapi.NewPhotoUpload(update.Message.Chat.ID, tgbotapi.FileBytes{Name: "loading.png", Bytes: pictures.loading}))
			msg.Text = "Попробуй /teach, /check или /author"
		}
		bot.Send(msg)
	}

	closeChan <- true
}
func processUser(update tgbotapi.Update) user {
	currentUser := user{}
	if users[update.Message.From.ID].authorized {
		currentUser = users[update.Message.From.ID]
		currentUser.applyUpdate(update)
		users[update.Message.From.ID] = currentUser
	} else {
		currentUser = user{
			update.Message.From.ID,
			update.Message.From.String(),
			fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName),
			"empty_string_currently",
			time.Now(),
			false,
			true,
		}
		users[update.Message.From.ID] = currentUser
	}

	return currentUser
}

// read config
func parseConfig() baseConfig {
	config := baseConfig{}
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(content, &config)

	return config
}

func initPictures() (pictures pictures) {
	var err error

	pictures.clever, err = ioutil.ReadFile(cleverImgPath)
	if err != nil {
		log.Fatal(err)
	}
	pictures.done, err = ioutil.ReadFile(doneImgPath)
	if err != nil {
		log.Fatal(err)
	}
	pictures.loading, err = ioutil.ReadFile(loadingImgPath)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func rebuild(branch string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	cmd := exec.Command(fmt.Sprintf("git checkout %s && git pull && go build", branch))
	err := cmd.Run()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
		bot.Send(msg)
	}
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Successfully switched to %s and updated", branch)))
	die("")
}

func die(sig interface{}) {
	os.Stdout.WriteString(fmt.Sprintf("Caught signal %s: shutting down.", sig))
	os.Exit(0)
}
