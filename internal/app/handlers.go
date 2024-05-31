package app

import (
	"cmd/task_bot/internal/app/utils"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"sync"
)

var taskID string
var taskInfo = map[int]struct {
	Description string
	URL         string
}{
	1: {"Follow Gummy on Twitter", "https://twitter.com/gummy"},
	2: {"Comment on YouTube for Gummy", "https://youtube.com/gummy"},
	3: {"Follow Baked on Twitter", "https://twitter.com/baked"},
	4: {"Comment on YouTube for Baked", "https://youtube.com/baked"},
}

type ChatState struct {
	sync.RWMutex
	M map[int64]string
}

func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, cState *ChatState) {
	userID := message.From.ID

	cState.RLock()
	state := cState.M[userID]
	cState.RUnlock()

	fmt.Println(cState.M)
	fmt.Println(userID)
	switch state {
	case "awaitingID":
		handleIDResponse(bot, message, cState)
	case "awaitingPhoto":
		handlePhotoUpload(bot, message, cState, taskID)
	default:
		if message.IsCommand() {
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
			case "profile":
				showUserProfile(bot, message.Chat.ID)
			default:
				log.Printf("Unknown command: %s", message.Text)
			}
		} else {
			log.Printf("Regular message received: %s", message.Text)
			handleRegularMessage(bot, message, cState)
		}
	}
}

func handleRegularMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, cState *ChatState) {
	userID := message.From.ID

	cState.RLock()
	state := cState.M[userID]
	cState.RUnlock()

	switch state {
	case "awaitingID":
		handleIDResponse(bot, message, cState)
	case "awaitingPhoto":
		handlePhotoUpload(bot, message, cState, taskID)
	default:
		log.Printf("No state to handle regular message: %s", message.Text)
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

func HandleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, cState *ChatState) {
	switch callback.Data {
	case "gummy":
		showGummyTasks(bot, callback.Message.Chat.ID)
	case "baked":
		showBakedTasks(bot, callback.Message.Chat.ID)
	case "submit_proof_gummy":
		showInitialPrompt(bot, callback.Message, cState, "gummy")
	case "submit_proof_baked":
		showInitialPrompt(bot, callback.Message, cState, "baked")
	default:
		log.Printf("Unknown callback: %s", callback.Data)
	}
}

func handleIDResponse(bot *tgbotapi.BotAPI, message *tgbotapi.Message, chatState *ChatState) {
	if message.Text == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "please send a valid ID")
		bot.Send(msg)
		return
	}

	fmt.Println("Received ID:", message.Text)

	taskID = message.Text
	// Set the state to await a photo
	chatState.Lock()
	chatState.M[message.From.ID] = "awaitingPhoto"
	chatState.Unlock()

	msg := tgbotapi.NewMessage(message.Chat.ID, "please submit a photo of proof")
	bot.Send(msg)
}

func showMemecoinOptions(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Choose a Memecoin:")
	msg.ReplyMarkup = memecoinKeyboard()
	bot.Send(msg)
}

func showUserProfile(bot *tgbotapi.BotAPI, chatID int64) {
	points, err := utils.GetUserPoints(chatID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Error fetching points"))
		return
	}

	var response string
	total := 0
	for task, pts := range points {
		total += pts
		response += fmt.Sprintf("%s: %d pts\n", task, pts)
	}

	if response == "" {
		response = "No points earned yet."
	}

	msg := tgbotapi.NewMessage(chatID, "*Total points earned: "+strconv.Itoa(total)+"*\n\n"+response)
	msg.ParseMode = "Markdown"
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
	twitterButton := tgbotapi.NewInlineKeyboardButtonURL("1. Follow $Gummy on Twitter", "https://twitter.com/gummy")
	youtubeButton := tgbotapi.NewInlineKeyboardButtonURL("2. Comment On Youtube", "https://youtube.com/gummy")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(twitterButton, youtubeButton),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Submit Proof of Completion", "submit_proof_gummy")),
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
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Submit Proof of Completion", "submit_proof_baked")),
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

func handlePhotoUpload(bot *tgbotapi.BotAPI, message *tgbotapi.Message, cState *ChatState, tID string) {
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

		fmt.Println("within the handle photo", tID)
		err = utils.CompleteTask(message.Chat.ID, tID, url)
		if err != nil {
			log.Fatal(err)
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, "*Proof uploaded successfully.*\nIn case of foul play we will penalize you")
		_, err = bot.Send(msg)
		if err != nil {
			log.Fatal(err)
		}

		profileButton := tgbotapi.NewKeyboardButton("Profile")
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(profileButton),
		)

		msg = tgbotapi.NewMessage(message.Chat.ID, "You can view your profile using the button below:")
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}

	cState.Lock()
	delete(cState.M, message.From.ID)
	cState.Unlock()
}

func showInitialPrompt(bot *tgbotapi.BotAPI, message *tgbotapi.Message, chatState *ChatState, memecoin string) {
	var msg tgbotapi.MessageConfig
	if memecoin == "gummy" {
		msg = tgbotapi.NewMessage(message.Chat.ID, "please send an id to submit your proof:\n1. Follow $Gummy on Twitter\n2. Comment On Youtube\n")
	} else if memecoin == "baked" {
		msg = tgbotapi.NewMessage(message.Chat.ID, "please send an id to submit your proof:\n3. Follow $Baked on Twitter\n4. Comment On Youtube\n")
	}

	bot.Send(msg)

	chatState.Lock()
	fmt.Println(message.Chat.ID)
	chatState.M[message.Chat.ID] = "awaitingID"
	chatState.Unlock()
}
