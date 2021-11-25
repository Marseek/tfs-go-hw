package repository

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

func (r *repo) WriteToTelegramBot(text string) {
	bot, err := tgbotapi.NewBotAPI("2146099871:AAF3T4RRFw6UhlG07i4e31O7grLwvuXXLH4")
	if err != nil {
		r.logger.Errorln("Can't connect to telegram: ", err)
		return
	}

	msg1 := tgbotapi.NewMessage(1689529148, text)
	_, err = bot.Send(msg1)
	if err != nil {
		r.logger.Errorln("Can't send message to telegram: ", err)
	}
}
