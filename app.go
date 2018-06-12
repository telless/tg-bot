package main

import (
	"os"
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"os/exec"
	"syscall"
	"os/signal"
	"encoding/json"
)

type Application struct {
	bot       *tgbotapi.BotAPI
	updates   tgbotapi.UpdatesChannel
	closeChan chan bool
	config    Config
	users     Users
	lessons   Lessons
}

type Config struct {
	Token    string `json:"token"`
	Domain   string `json:"domain"`
	RootPass string `json:"root_pass"`
}

// process messages
func (app *Application) run() {
	// listen for OS signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func(c chan os.Signal) {
		sig := <-c
		app.stop(sig)
	}(signals)

	for update := range app.updates {
		user := app.users.findOrCreateUser(update)
		os.Stdout.WriteString(fmt.Sprintf("Get message %s from %s\n", update.Message.Text, update.Message.From.String()))
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
				msg.Text = "Арсений Скурт @skurtars"
			case "auth": // some easter egg here
				if update.Message.CommandArguments() == app.config.RootPass {
					user.HasAdminRights = true
					app.users.Users[user.Id] = user
					msg.Text = fmt.Sprintf("Привет %s (%s), ты теперь админ!", user.FullName, user.Username)
				}
			case "build":
				if user.HasAdminRights {
					if update.Message.CommandArguments() != "" {
						app.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Пытаюсь пересобраться"))
						branch := update.Message.CommandArguments()
						err := app.rebuild(branch)
						if err != nil {
							app.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
						} else {
							app.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Успешно обновил ветку %s и пересобрался", branch)))
							app.stop("Rebuild")
						}
					} else {
						msg.Text = "Укажите имя ветки для пересборки"
					}
				}
			case "whoami":
				msg.Text = fmt.Sprintf("Привет %s (%s)", user.FullName, user.Username)

			case "add_lesson":
				if user.HasAdminRights {
					if update.Message.CommandArguments() != "" {
						app.lessons.add(update.Message.CommandArguments())
					} else {
						msg.Text = "Введите корректный JSON со структурой урока"
					}
				}

			case "print_lessons":
				if user.HasAdminRights {
					lessonsJson, _ := json.Marshal(app.lessons)
					msg.Text = string(lessonsJson)
				}

			default:
				msg.Text = "Попробуй /teach, /check или /author"
			}
		} else {
			msg.Text = "Попробуй /teach, /check или /author"
		}
		if msg.Text != "" {
			app.bot.Send(msg)
		}
	}

	app.closeChan <- true
}

func (app *Application) rebuild(branch string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("git fetch --all && git checkout %s && git pull && go build", branch))
	err := cmd.Run()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return err
	} else {
		return nil
	}
}

func (app *Application) stop(sig interface{}) {
	saveUsers(app.users)
	saveLessons(app.lessons)
	os.Stdout.WriteString(fmt.Sprintf("Caught signal %s: shutting down.\n", sig))
	os.Exit(0)
}
