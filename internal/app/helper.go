package app

import (
	"cmd/task_bot/internal/app/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func ShowTasks(bot *tgbotapi.BotAPI, chatID int64, offsetMap *map[int64]int64, messageID *map[int64]int, tID *map[int64]int) {
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
	(*tID)[chatID] = task.ID
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

func ShowSubTaskPrevTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int, tid *map[int64]int) {
	taskId := (*tid)[chatID]
	taskCount, err := utils.GetTotalNumberOfSubTasks(chatID, taskId)
	if err != nil {
		log.Fatal("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] - 1 + *taskCount) % *taskCount

	HandleTakeTask(bot, chatID, taskId, userOffsets, mid)
}

func ShowNextTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int) {
	taskCount, err := utils.GetTotalNumberOfTasks(chatID)
	if err != nil {
		fmt.Println("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] + 1) % *taskCount

	EditTaskMessage(bot, chatID, userOffsets, mid)
}

func ShowSubTaskNextTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int, tid *map[int64]int) {
	taskId := (*tid)[chatID]

	taskCount, err := utils.GetTotalNumberOfSubTasks(chatID, taskId)
	if err != nil {
		fmt.Println("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] + 1) % *taskCount

	HandleTakeTask(bot, chatID, taskId, userOffsets, mid)
}

func HandleTakeTask(bot *tgbotapi.BotAPI, chatID int64, tid int, offsetMap *map[int64]int64, messageID *map[int64]int) {
	offset := (*offsetMap)[chatID]

	if (*offsetMap) == nil {
		offset = 0
	}

	mid := (*messageID)[chatID]

	subTasks, err := utils.GetUnCompletedSubTasks(chatID, tid, offset)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	msgText := fmt.Sprintf("Task Name: %s \nTask Description: %s \nTask Point: %d", subTasks.Name, subTasks.Descriptions, subTasks.Points)

	editMsg := tgbotapi.NewEditMessageText(chatID, mid, msgText)

	prevBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Prev", "sub_task_prev")
	nextBtn := tgbotapi.NewInlineKeyboardButtonData("Next ➡️", "sub_task_next")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(prevBtn, nextBtn),
	)

	if subTasks.Links != nil {
		subTaskBtn := tgbotapi.NewInlineKeyboardButtonURL(subTasks.Name, *subTasks.Links)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(subTaskBtn))
	}

	editMsg.ReplyMarkup = &keyboard
	editMsg.ParseMode = "Markdown"
	sentMsg, err := bot.Send(editMsg)

	if err != nil {
		log.Fatal("error", err.Error())
	}

	(*messageID)[chatID] = sentMsg.MessageID
}
