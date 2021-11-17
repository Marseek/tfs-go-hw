package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/Marseek/tfs-go-hw/home_work_04/domain"
)

var mu sync.Mutex

func (storage *Repo) ShowAll() []domain.Messages {
	mu.Lock()
	m := storage.ToAllMessages
	mu.Unlock()
	return m
}

func (storage *Repo) SendToAll(mes domain.Messages) {
	mu.Lock()
	storage.ToAllMessages = append(storage.ToAllMessages, mes)
	mu.Unlock()
}

func (storage *Repo) ShowMy(user string) string {
	outString := fmt.Sprintf("%s's messages\n", user)
	mu.Lock()
	m := storage.PersonalMessages[user]
	mu.Unlock()
	for _, val := range m {
		outString += fmt.Sprintf("%-14s:%-50s: %s\n", val.From, val.Message, val.Time.Format("15:04:05"))
	}
	return outString
}

func (storage *Repo) SendToSmb(mes domain.Messages, user string) {
	mu.Lock()
	storage.PersonalMessages[user] = append(storage.PersonalMessages[user], domain.Messages{From: mes.From, Message: mes.Message, Time: time.Now()})
	mu.Unlock()
}

func (storage *Repo) RegisterUser(name, pass string) {
	mu.Lock()
	storage.UserNamesAndPass[name] = pass
	mu.Unlock()
}

func (storage *Repo) UserIsPresent(to string) bool {
	mu.Lock()
	_, ok := storage.UserNamesAndPass[to]
	mu.Unlock()
	return ok
}
