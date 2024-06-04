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

var userOffsets = make(map[int64]int64)
var mID = make(map[int64]int)
var tID = make(map[int64]int)
var stOffsets = make(map[int64]int64)

type ChatState struct {
	sync.RWMutex
	M map[int64]string
}

func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, cState *ChatState) {
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
			case "profile":
				showUserProfile(bot, message.Chat.ID)
			case "help":
				showHow(bot, message.Chat.ID)
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
	photo := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileURL("https://gummyonsol.com/images/529376304672a8a43191f520936dbd48.png"))
	_, err := bot.Send(photo)
	if err != nil {
		log.Panic(err)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "*Welcome to Fefe bot! 🤖* \n\n This bot lets you earn rewards by completing simple tasks. \n\n - ✨ Please choose from the button below")
	msg.ParseMode = "Markdown"
	howItWorksBtn := tgbotapi.NewInlineKeyboardButtonData("❓How it works", "help")
	earnBtn := tgbotapi.NewInlineKeyboardButtonData("💰 Earn", "earn")
	adsBtn := tgbotapi.NewInlineKeyboardButtonData("💻 Advertise", "ads")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(howItWorksBtn),
		tgbotapi.NewInlineKeyboardRow(earnBtn),
		tgbotapi.NewInlineKeyboardRow(adsBtn),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

//func mainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
//	memecoinButton := tgbotapi.NewKeyboardButton("Memecoin")
//	helpButton := tgbotapi.NewKeyboardButton("Help")
//
//	keyboard := tgbotapi.NewReplyKeyboard(
//		tgbotapi.NewKeyboardButtonRow(memecoinButton, helpButton),
//	)
//a
//	return keyboard
//}

func HandleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, cState *ChatState) {
	tid := tID[callback.Message.Chat.ID]

	switch callback.Data {
	case "help":
		showHow(bot, callback.Message.Chat.ID)
	case "earn":
		ShowTasks(bot, callback.Message.Chat.ID, &userOffsets, &mID, &tID)
	case "prev":
		ShowPrevTask(bot, callback.Message.Chat.ID, &userOffsets, &mID)
	case "next":
		ShowNextTask(bot, callback.Message.Chat.ID, &userOffsets, &mID)
	case "sub_task_prev":
		ShowSubTaskPrevTask(bot, callback.Message.Chat.ID, &stOffsets, &mID, &tID)
	case "sub_task_next":
		ShowSubTaskNextTask(bot, callback.Message.Chat.ID, &stOffsets, &mID, &tID)
	case "complete_task_" + strconv.Itoa(tid):
		HandleTakeTask(bot, callback.Message.Chat.ID, tid, &stOffsets, &mID)
	case "profile":
		showUserProfile(bot, callback.Message.Chat.ID)
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

	taskID = message.Text
	chatState.Lock()
	chatState.M[message.From.ID] = "awaitingPhoto"
	chatState.Unlock()

	msg := tgbotapi.NewMessage(message.Chat.ID, "please submit a photo of proof")
	bot.Send(msg)
}

//func showMemecoinOptions(bot *tgbotapi.BotAPI, chatID int64) {
//	//photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL("https://mybwvaraeybzsepptucn.supabase.co/storage/v1/object/public/task_memecoin/DALL_E_May_31_Gummy_Bear_Memecoins.webp"))
//	//_, err := bot.Send(photo)
//	//if err != nil {
//	//	log.Panic(err)
//	//}
//
//	msg := tgbotapi.NewMessage(chatID, "![choose memecoin](https://mybwvaraeybzsepptucn.supabase.co/storage/v1/object/public/task_memecoin/DALL_E_May_31_Gummy_Bear_Memecoins.webp) Choose a *Memecoin*: ")
//	msg.ParseMode = "Markdown"
//	msg.ReplyMarkup = memecoinKeyboard()
//	bot.Send(msg)
//}

func showHow(bot *tgbotapi.BotAPI, chatID int64) {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL("https://gummyonsol.com/images/529376304672a8a43191f520936dbd48.png"))
	_, err := bot.Send(photo)
	if err != nil {
		log.Panic(err)
	}

	msg := tgbotapi.NewMessage(chatID, "*❓ How it works❓* \n\nFefe bot is a cryptocurrency-based community task platform. \n\nby using this bot, you agree to Terms of Services and Privacy Policy. \n\nHere are all my commands: \n\n/start - show the main menu \n/earn - start completing task and earn points\n/balance - show your balance\n/help - Show help. \n\n *Start using fefe bot and earn points 🏆*")
	msg.ParseMode = "Markdown"
	//howItWorksBtn := tgbotapi.NewInlineKeyboardButtonData("❓How it works", "help")
	//earnBtn := tgbotapi.NewInlineKeyboardButtonData("💰 Earn", "earn")
	//adsBtn := tgbotapi.NewInlineKeyboardButtonData("💻 Advertise", "ads")
	//
	//keyboard := tgbotapi.NewInlineKeyboardMarkup(
	//	tgbotapi.NewInlineKeyboardRow(howItWorksBtn),
	//	tgbotapi.NewInlineKeyboardRow(earnBtn),
	//	tgbotapi.NewInlineKeyboardRow(adsBtn),
	//)
	//msg.ReplyMarkup = keyboard
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

	msg := tgbotapi.NewMessage(chatID, "![Points](https://pbs.twimg.com/media/GN28dBfX0AA2dt-?format=jpg&name=large) \n*Total points earned: "+strconv.Itoa(total)+"*\n\n"+response)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func memecoinKeyboard() tgbotapi.InlineKeyboardMarkup {
	gummyButton := tgbotapi.NewInlineKeyboardButtonData("$Gummy", "gummy")
	bakedButton := tgbotapi.NewInlineKeyboardButtonData("$Baked", "baked")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(gummyButton, bakedButton),
	)

	return keyboard
}

func showGummyTasks(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Choose gummy tasks to [earn points](https://gummyonsol.com/images/f0d9f977ea430a9b57a7d4f7277df4eb.png):")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = gummyTaskKeyboard()
	bot.Send(msg)
}

func gummyTaskKeyboard() tgbotapi.InlineKeyboardMarkup {
	twitterButton := tgbotapi.NewInlineKeyboardButtonURL("1. Follow $Gummy on Twitter", "https://x.com/gummyonsolana")
	youtubeButton := tgbotapi.NewInlineKeyboardButtonURL("2. Comment On Youtube", "https://www.youtube.com/watch?v=XKB7EWvocEo")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(twitterButton, youtubeButton),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Submit Proof of Completion", "submit_proof_gummy")),
	)

	return keyboard
}

func showBakedTasks(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Choose Baked tasks to ![earn points](https://memestaking.com/_nuxt/baked-d.VVFJ9SFz.png):")
	msg.ReplyMarkup = bakedTaskKeyboard()
	bot.Send(msg)
}

func bakedTaskKeyboard() tgbotapi.InlineKeyboardMarkup {
	twitterButton := tgbotapi.NewInlineKeyboardButtonURL("3. Follow $Baked Twitter", "https://x.com/bakedtoken")
	youtubeButton := tgbotapi.NewInlineKeyboardButtonURL("4. Comment On Youtube", "https://www.youtube.com/watch?v=XKB7EWvocEo")

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
		//showMemecoinOptions(bot, message.Chat.ID)
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

		profileButton := tgbotapi.NewInlineKeyboardButtonData("profile", "profile")

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(profileButton),
		)

		msg = tgbotapi.NewMessage(message.Chat.ID, "You can view your profile using the button below:")
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}

	cState.Lock()
	delete(cState.M, message.From.ID)
	cState.Unlock()
}

//func showInitialPrompt(bot *tgbotapi.BotAPI, message *tgbotapi.Message, chatState *ChatState, memecoin string) {
//	var msg tgbotapi.MessageConfig
//	if memecoin == "gummy" {
//		msg = tgbotapi.NewMessage(message.Chat.ID, "please send an id to submit your proof:\n1. Follow $Gummy on Twitter\n2. Comment On Youtube\n")
//	} else if memecoin == "baked" {
//		msg = tgbotapi.NewMessage(message.Chat.ID, "please send an id to submit your proof:\n3. Follow $Baked on Twitter\n4. Comment On Youtube\n")
//	}
//
//	bot.Send(msg)
//
//	chatState.Lock()
//	fmt.Println(message.Chat.ID)
//	chatState.M[message.Chat.ID] = "awaitingID"
//	chatState.Unlock()
//}
