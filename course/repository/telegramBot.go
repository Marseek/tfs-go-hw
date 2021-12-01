package repository

import (
	"context"
	"course/pkg/telegrampb"
)

func (r *Repo) WriteToTelegramBot(text string) {
	_, _ = r.tgClient.SendMessage(context.Background(), &telegrampb.Request{Req: text})
}
