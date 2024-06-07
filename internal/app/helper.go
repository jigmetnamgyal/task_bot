package app

import (
	"cmd/task_bot/internal/app/utils"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

func ShowTasks(bot *tgbotapi.BotAPI, chatID int64, offsetMap *map[int64]int64, messageID *map[int64]int, tID *map[int64]int) {
	offset := (*offsetMap)[chatID]

	if (*offsetMap) == nil {
		offset = 0
	}

	count, err := utils.GetTotalNumberOfTasks(chatID)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if *count > 0 {
		task, err := utils.GetUnCompletedTasks(chatID, offset)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		msgText := fmt.Sprintf("Task Name: %s \nTask Description: %s", task.Name, task.Descriptions)

		msg := tgbotapi.NewMessage(chatID, msgText)

		keyboard := tgbotapi.NewInlineKeyboardMarkup()
		fmt.Println("task count", *count)
		if *count > 1 {
			prevBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Prev", fmt.Sprintf("prev"))
			nextBtn := tgbotapi.NewInlineKeyboardButtonData("Next ➡️", fmt.Sprintf("next"))

			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(prevBtn, nextBtn))
		}

		skipBtn := tgbotapi.NewInlineKeyboardButtonData("❌️ Skip Task", "skip_task")

		takeBtn := tgbotapi.NewInlineKeyboardButtonData("🤑 Take", fmt.Sprintf("complete_task_%d", task.ID))

		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(skipBtn, takeBtn))

		msg.ReplyMarkup = keyboard
		msg.ParseMode = "Markdown"
		sentMsg, err := bot.Send(msg)

		if err != nil {
			log.Fatal("error", err.Error())
		}

		fmt.Println("From show tasks", task.ID)
		(*messageID)[chatID] = sentMsg.MessageID
		(*tID)[chatID] = task.ID
	} else {
		msgText := fmt.Sprintf("You have completed all the tasks, you will be notified when there is new tasks")

		editMsg := tgbotapi.NewMessage(chatID, msgText)
		backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", fmt.Sprintf("back"))

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(backBtn),
		)

		editMsg.ReplyMarkup = &keyboard
		editMsg.ParseMode = "Markdown"
		bot.Send(editMsg)
	}

}

func EditTaskMessage(bot *tgbotapi.BotAPI, chatID int64, offsetMap *map[int64]int64, messageID *map[int64]int, tID *map[int64]int) {
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

	count, err := utils.GetTotalNumberOfTasks(chatID)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if *count > 0 {
		msgText := fmt.Sprintf("Task Name: %s \nTask Description: %s", task.Name, task.Descriptions)

		editMsg := tgbotapi.NewEditMessageText(chatID, mid, msgText)

		keyboard := tgbotapi.NewInlineKeyboardMarkup()
		if *count > 1 {
			prevBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Prev", fmt.Sprintf("prev"))
			nextBtn := tgbotapi.NewInlineKeyboardButtonData("Next ➡️", fmt.Sprintf("next"))

			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(prevBtn, nextBtn))
		}

		skipBtn := tgbotapi.NewInlineKeyboardButtonData("❌️ Skip Task", "skip_task")
		takeBtn := tgbotapi.NewInlineKeyboardButtonData("🤑 Take", fmt.Sprintf("complete_task_%d", task.ID))

		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(takeBtn, skipBtn))

		editMsg.ReplyMarkup = &keyboard
		editMsg.ParseMode = "Markdown"
		sentMsg, err := bot.Send(editMsg)

		if err != nil {
			log.Fatal("error", err.Error())
		}

		(*messageID)[chatID] = sentMsg.MessageID
		(*tID)[chatID] = task.ID
	} else {
		msgText := fmt.Sprintf("You have completed all the tasks, you will be notified when there is new tasks")

		editMsg := tgbotapi.NewEditMessageText(chatID, mid, msgText)
		backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", fmt.Sprintf("back"))

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(backBtn),
		)

		editMsg.ReplyMarkup = &keyboard
		editMsg.ParseMode = "Markdown"
		bot.Send(editMsg)
	}

}

func ShowPrevTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int, tID *map[int64]int) {
	taskCount, err := utils.GetTotalNumberOfTasks(chatID)
	if err != nil {
		log.Fatal("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] - 1 + *taskCount) % *taskCount

	EditTaskMessage(bot, chatID, userOffsets, mid, tID)
}

func ShowSubTaskPrevTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int, tid *map[int64]int, stid *map[int64]int, cancleFuncs map[int64]context.CancelFunc) {
	taskId := (*tid)[chatID]
	taskCount, err := utils.GetTotalNumberOfSubTasks(chatID, taskId)
	if err != nil {
		log.Fatal("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] - 1 + *taskCount) % *taskCount

	HandleTakeTask(bot, chatID, taskId, userOffsets, mid, stid, cancleFuncs)
}

func ShowNextTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int, tID *map[int64]int) {
	taskCount, err := utils.GetTotalNumberOfTasks(chatID)
	if err != nil {
		fmt.Println("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] + 1) % *taskCount

	EditTaskMessage(bot, chatID, userOffsets, mid, tID)
}

func ShowSubTaskNextTask(bot *tgbotapi.BotAPI, chatID int64, userOffsets *map[int64]int64, mid *map[int64]int, tid *map[int64]int, stid *map[int64]int, cancleFuncs map[int64]context.CancelFunc) {
	taskId := (*tid)[chatID]

	taskCount, err := utils.GetTotalNumberOfSubTasks(chatID, taskId)
	if err != nil {
		fmt.Println("error", err.Error())
	}

	(*userOffsets)[chatID] = ((*userOffsets)[chatID] + 1) % *taskCount

	HandleTakeTask(bot, chatID, taskId, userOffsets, mid, stid, cancleFuncs)
}

// HandleTakeTask TODO: subtask server dead and come back it divides by 0 prevent
func HandleTakeTask(bot *tgbotapi.BotAPI, chatID int64, tid int, offsetMap *map[int64]int64, messageID *map[int64]int, stid *map[int64]int, cancelFuncs map[int64]context.CancelFunc) {
	offset := (*offsetMap)[chatID]

	if (*offsetMap) == nil {
		offset = 0
	}

	mid := (*messageID)[chatID]

	count, err := utils.GetTotalNumberOfSubTasks(chatID, tid)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if *count > 0 {
		subTasks, err := utils.GetUnCompletedSubTasks(chatID, tid, offset)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		msgText := fmt.Sprintf("Task Name: %s \n\nTask Description: %s \n\nTask Point: *%d*", subTasks.Name, subTasks.Descriptions, subTasks.Points)

		editMsg := tgbotapi.NewEditMessageText(chatID, mid, msgText)

		keyboard := tgbotapi.NewInlineKeyboardMarkup()
		if *count > 1 {
			prevBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Prev", "sub_task_prev")
			nextBtn := tgbotapi.NewInlineKeyboardButtonData("Next ➡️", "sub_task_next")

			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(prevBtn, nextBtn))
		}

		skipBtn := tgbotapi.NewInlineKeyboardButtonData("❌️ Skip Task", "skip_task")
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(skipBtn))

		if subTasks.Links != nil {
			subTaskBtn := tgbotapi.NewInlineKeyboardButtonURL(subTasks.Name, *subTasks.Links)
			confirmBtn := tgbotapi.NewInlineKeyboardButtonData("✅ confirm task", fmt.Sprintf("confirm_subtask_%d", subTasks.ID))

			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(subTaskBtn))

			// Cancel the previous context if it exists
			if cancelFunc, exists := cancelFuncs[chatID]; exists {
				cancelFunc()
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancelFuncs[chatID] = cancel

			go func() {
				select {
				case <-ctx.Done():
					// Context cancelled, exit goroutine
					return
				case <-time.After(5 * time.Second):
					// Time elapsed, proceed with the task
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(confirmBtn))

					editMsg.ReplyMarkup = &keyboard
					editMsg.ParseMode = "Markdown"
					sentMsg, err := bot.Send(editMsg)

					if err != nil {
						log.Fatal("error", err.Error())
					}

					(*messageID)[chatID] = sentMsg.MessageID
					(*stid)[chatID] = subTasks.ID
				}
			}()
		}

		editMsg.ReplyMarkup = &keyboard
		editMsg.ParseMode = "Markdown"
		sentMsg, err := bot.Send(editMsg)

		if err != nil {
			log.Fatal("error", err.Error())
		}

		(*messageID)[chatID] = sentMsg.MessageID
		(*stid)[chatID] = subTasks.ID
	} else {
		msgText := fmt.Sprintf("You have completed all the tasks, you will be notified when there is new tasks")

		editMsg := tgbotapi.NewMessage(chatID, msgText)
		backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", fmt.Sprintf("back"))

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(backBtn),
		)

		editMsg.ReplyMarkup = &keyboard
		editMsg.ParseMode = "Markdown"
		bot.Send(editMsg)
	}
}

func handleTaskComplete(bot *tgbotapi.BotAPI, chatID int64, stid int) {
	var userID int
	err := utils.DB.QueryRow("SELECT id FROM users WHERE telegram_id = $1", chatID).Scan(&userID)
	if err != nil {
		log.Fatal(err.Error())
	}

	queryString := `INSERT INTO user_sub_tasks (user_id, sub_task_id, completed) VALUES ($1, $2, $3)`

	prepare, err := utils.DB.Prepare(queryString)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = prepare.Exec(userID, stid, true)
	if err != nil {
		log.Fatal(err.Error())
	}

	EditTaskMessage(bot, chatID, &userOffsets, &mID, &tID)
}
