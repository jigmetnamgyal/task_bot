package app

import (
	"cmd/task_bot/internal/app/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func ShowTasks(bot *tgbotapi.BotAPI, chatID int64, offsetMap *map[int64]int64, messageID *map[int64]int) {
	offset := (*offsetMap)[chatID]

	if (*offsetMap) == nil {
		offset = 0
	}

	task, err := utils.GetUnCompletedTasks(chatID, offset)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	msgText := fmt.Sprintf("Task Name: %s \nTask Description: %s", task.Name, task.Descriptions)

	msg := tgbotapi.NewMessage(chatID, msgText)

	prevBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Prev", fmt.Sprintf("prev"))
	takeBtn := tgbotapi.NewInlineKeyboardButtonData("🤑 Take", fmt.Sprintf("complete_task_%d", task.ID))
	nextBtn := tgbotapi.NewInlineKeyboardButtonData("Next ➡️", fmt.Sprintf("next"))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(prevBtn, takeBtn, nextBtn),
	)

	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	sentMsg, err := bot.Send(msg)

	if err != nil {
		log.Fatal("error", err.Error())
	}

	(*messageID)[chatID] = sentMsg.MessageID
}

func EditTaskMessage(bot *tgbotapi.BotAPI, chatID int64, offsetMap *map[int64]int64, messageID *map[int64]int) {
	offset := (*offsetMap)[chatID]

	if (*offsetMap) == nil {
		offset = 0
	}

	mid := (*messageID)[chatID]

	task, err := utils.GetUnCompletedTasks(chatID, offset)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	msgText := fmt.Sprintf("Task Name: %s \nTask Description: %s", task.Name, task.Descriptions)

	editMsg := tgbotapi.NewEditMessageText(chatID, mid, msgText)

	prevBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Prev", fmt.Sprintf("prev"))
	takeBtn := tgbotapi.NewInlineKeyboardButtonData("🤑 Take", fmt.Sprintf("complete_task_%d", task.ID))
	nextBtn := tgbotapi.NewInlineKeyboardButtonData("Next ➡️", fmt.Sprintf("next"))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(prevBtn, takeBtn, nextBtn),
	)

	editMsg.ReplyMarkup = &keyboard
	editMsg.ParseMode = "Markdown"
	sentMsg, err := bot.Send(editMsg)

	if err != nil {
		log.Fatal("error", err.Error())
	}

	(*messageID)[chatID] = sentMsg.MessageID
}

func ShowPrevTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int) {
	taskCount, err := utils.GetTotalNumberOfTasks(chatID)
	if err != nil {
		log.Fatal("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] - 1 + *taskCount) % *taskCount

	EditTaskMessage(bot, chatID, userOffsets, mid)
}

func ShowNextTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int) {
	taskCount, err := utils.GetTotalNumberOfTasks(chatID)
	if err != nil {
		fmt.Println("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] + 1) % *taskCount

	EditTaskMessage(bot, chatID, userOffsets, mid)
	//ShowTasks(bot, chatID, userOffsets)
}
