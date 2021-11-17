package repository

import "github.com/Marseek/tfs-go-hw/home_work_04/domain"

type Repo struct {
	*domain.MessageStorage
}

func NewRepository(mes *domain.MessageStorage) Repository {
	return &Repo{
		mes,
	}
}

type Repository interface {
	RegisterUser(name, pass string)
	UserIsPresent(to string) bool
	SendToAll(mes domain.Messages)
	SendToSmb(mes domain.Messages, user string)
	ShowAll() []domain.Messages
	ShowMy(string) string
}
