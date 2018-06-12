package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"time"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"os"
)

const usersInitFile = "users.json"

type Users struct {
	Users map[int]User `json:"users"`
}

type User struct {
	Id               int           `json:"id"`
	Username         string        `json:"username"`
	FullName         string        `json:"full_name"`
	LastVisit        time.Time     `json:"last_visit"`
	HasAdminRights   bool          `json:"has_admin_rights"`
	Authorized       bool          `json:"authorized"`
	CurrentLesson    CurrentLesson `json:"current_lesson"`
	CompletedLessons []int         `json:"completed_lessons"`
}

type CurrentLesson struct {
	LessonId int `json:"lesson_id"`
	PageId   int `json:"page_id"`
}

func initUsers() (users Users) {
	userJson, err := ioutil.ReadFile(usersInitFile)
	if err != nil {
		return Users{make(map[int]User)}
	}

	json.Unmarshal(userJson, &users)

	return
}

func saveUsers(users Users) {
	file, err := os.OpenFile(usersInitFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	logError(err)
	defer file.Close()
	data, err := json.Marshal(users)
	logError(err)
	file.Write(data)
	file.WriteString("\n")
	os.Stdout.WriteString("Users saved\n")
}

func (u *User) applyUpdate(update tgbotapi.Update) {
	u.LastVisit = time.Now()
	u.Username = update.Message.From.String()
	u.FullName = fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
}

func (users *Users) findOrCreateUser(update tgbotapi.Update) (currentUser User) {
	if users.Users[update.Message.From.ID].Authorized {
		currentUser = users.Users[update.Message.From.ID]
		currentUser.applyUpdate(update)
	} else {
		currentUser = User{
			Id:         update.Message.From.ID,
			Username:   update.Message.From.String(),
			FullName:   fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName),
			LastVisit:  time.Now(),
			Authorized: true,
		}
	}
	users.Users[update.Message.From.ID] = currentUser

	return
}
