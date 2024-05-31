package main

import (
	"cmd/task_bot/internal/app"
	"cmd/task_bot/internal/app/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func init() {
	utils.LoadEnvironmentVariable()
	utils.ConnectToDb()
}

type BotState int

var chatState = make(map[int64]BotState)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Photo != nil {
				app.HandlePhotoUpload(bot, update.Message)
			} else {
				app.HandleMessage(bot, update.Message)
			}
		} else if update.CallbackQuery != nil {
			app.HandleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}
