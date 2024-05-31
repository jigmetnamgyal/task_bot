package app

import (
	"cmd/task_bot/internal/app/utils"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

var taskInfo = map[int]struct {
	Description string
	URL         string
}{
	1: {"Follow Gummy on Twitter", "https://twitter.com/gummy"},
	2: {"Comment on YouTube for Gummy", "https://youtube.com/gummy"},
	3: {"Follow Baked on Twitter", "https://twitter.com/baked"},
	4: {"Comment on YouTube for Baked", "https://youtube.com/baked"},
}

func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		exists, err := userExists(message.From.ID)
		if err != nil {
			log.Printf("Failed to check user exist")
			return
		}

		if !exists {
			err = utils.AddUser(message.From.ID)
			if err != nil {
				log.Printf("Failed to create user")
				return
			}
		}

		handleStartCommand(bot, message)
	case "memecoin":
		showMemecoinOptions(bot, message.Chat.ID)
	default:
		log.Printf("Unknown command: %s", message.Text)
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome to the Memecoin Bot!")
	msg.ReplyMarkup = mainMenuKeyboard()
	bot.Send(msg)
}

func mainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	memecoinButton := tgbotapi.NewKeyboardButton("Memecoin")
	helpButton := tgbotapi.NewKeyboardButton("Help")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(memecoinButton, helpButton),
	)

	return keyboard
}

func HandleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	switch callback.Data {
	//case "memecoin":
	//	showMemecoinOptions(bot, callback)
	case "gummy":
		showGummyTasks(bot, callback.Message.Chat.ID)
	case "baked":
		showBakedTasks(bot, callback.Message.Chat.ID)
	case "submit_proof":
		handlePhotoUpload(bot, callback.Message)
	default:
		log.Printf("Unknown callback: %s", callback.Data)
	}
}

// "INSERT INTO user_tasks (user_id, task_id, completed) VALUES ($1, $2, TRUE) ON CONFLICT (user_id, task_id) DO NOTHING"
func showMemecoinOptions(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Choose a Memecoin:")
	msg.ReplyMarkup = memecoinKeyboard()
	bot.Send(msg)
}

func memecoinKeyboard() tgbotapi.InlineKeyboardMarkup {
	gummyButton := tgbotapi.NewInlineKeyboardButtonData("Gummy", "gummy")
	bakedButton := tgbotapi.NewInlineKeyboardButtonData("Baked", "baked")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(gummyButton, bakedButton),
	)

	return keyboard
}

func showGummyTasks(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Choose gummy tasks to earn points:")
	msg.ReplyMarkup = gummyTaskKeyboard()
	bot.Send(msg)
}

func gummyTaskKeyboard() tgbotapi.InlineKeyboardMarkup {
	twitterButton := tgbotapi.NewInlineKeyboardButtonURL("1. Follow $Gummy on Twitter", "https://twitter.com/baked")
	youtubeButton := tgbotapi.NewInlineKeyboardButtonURL("2. Comment On Youtube", "https://youtube.com/baked")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(twitterButton, youtubeButton),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Submit Proof of Completion", "submit_proof")),
	)

	return keyboard
}

func showBakedTasks(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Choose Baked tasks to earn points")
	msg.ReplyMarkup = bakedTaskKeyboard()
	bot.Send(msg)
}

func bakedTaskKeyboard() tgbotapi.InlineKeyboardMarkup {
	twitterButton := tgbotapi.NewInlineKeyboardButtonURL("3. Follow $Baked Twitter", "https://twitter.com/baked")
	youtubeButton := tgbotapi.NewInlineKeyboardButtonURL("4. Comment On Youtube", "https://youtube.com/baked")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(twitterButton, youtubeButton),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Submit Proof of Completion", "submit_proof")),
	)

	return keyboard
}

func userExists(telegramID int64) (bool, error) {
	var count int64

	prepare, err := utils.DB.Prepare("select COUNT(*) FROM users WHERE telegram_id = ($1)")
	if err != nil {
		return false, err
	}

	err = prepare.QueryRow(telegramID).Scan(&count)
	if err != nil {
		return false, err
	}

	condition := count == 1
	return condition, nil
}

func handlePhotoUpload(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	if message.Photo == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "please submit a file")
		bot.Send(msg)
		showMemecoinOptions(bot, message.Chat.ID)
	} else {
		fileID := message.Photo[len(message.Photo)-1].FileID

		url, err := bot.GetFileDirectURL(fileID)
		if err != nil {
			log.Printf("Failed to get file URL: %s", err)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Failed to upload proof."))
			return
		}

		tID := message.Text

		fmt.Println("task id ya", tID)

		err = utils.CompleteTask(message.From.ID, "", url)
		if err != nil {
			log.Fatal(err)
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, "*Proof uploaded successfully.*\nIn case of cheating we will penalize you")
		_, err = bot.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
