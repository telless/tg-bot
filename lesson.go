package main

import (
	"io/ioutil"
	"encoding/json"
	"os"
)

const lessonsInitFile = "lessons.json";

type Lessons struct {
	LastId  int            `json:"last_id"`
	Lessons map[int]Lesson `json:"lessons"`
}

type Lesson struct {
	Id       int            `json:"id"`
	Pages    map[int]string `json:"pages"`
	Question Question       `json:"question"`
}

type Question struct {
	Text    string         `json:"text"`
	Answers map[int]Answer `json:"answers"`
}

type Answer struct {
	Id        int    `json:"id"`
	IsCorrect bool   `json:"is_correct"`
	Text      string `json:"text"`
}

func (lessons *Lessons) add(jsonData string) error {
	id := lessons.LastId
	lesson := Lesson{Id: id}
	err := json.Unmarshal([]byte(jsonData), &lesson)
	if err == nil {
		lessons.LastId = id + 1
		lessons.Lessons[id] = lesson
	}

	return err
}

func initLessons() (lessons Lessons) {
	content, err := ioutil.ReadFile(lessonsInitFile)
	if err != nil {
		return Lessons{0, make(map[int]Lesson)}
	}

	json.Unmarshal(content, &lessons)

	return
}

func saveLessons(lessons Lessons) {
	file, err := os.OpenFile(lessonsInitFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	logError(err)
	defer file.Close()
	data, err := json.Marshal(lessons)
	logError(err)
	file.Write(data)
	file.WriteString("\n")
	os.Stdout.WriteString("Lessons saved\n")
}
