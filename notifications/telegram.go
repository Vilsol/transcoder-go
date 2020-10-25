package notifications

import (
	"fmt"
	"github.com/Vilsol/transcoder-go/models"
	"github.com/Vilsol/transcoder-go/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

var tgBot *tgbotapi.BotAPI

var messageKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Skip", "skip"),
	),
)

func init() {
	initialize = append(initialize, func() {
		if viper.GetString("tg-bot-key") != "" && viper.GetString("tg-chat-id") != "" {
			var err error
			tgBot, err = tgbotapi.NewBotAPI(viper.GetString("tg-bot-key"))

			if err != nil {
				log.Fatalf("Error initializing telegram bot: %s", err)
				return
			}

			log.Printf("Telegram connected: %s", tgBot.Self.UserName)

			var currentMessage *tgbotapi.Message

			lastMessage := int64(0)

			chatIDStr := viper.GetString("tg-chat-id")

			chatID, err := strconv.ParseInt(chatIDStr, 10, 64)

			if err != nil {
				chat, err := tgBot.GetChat(tgbotapi.ChatConfig{
					SuperGroupUsername: chatIDStr,
				})

				if err != nil {
					log.Fatalf("Chat not found: %s", chatIDStr)
					return
				}

				chatID = chat.ID
			}

			if viper.GetInt("tg-admin-id") != 0 {
				go func() {
					u := tgbotapi.NewUpdate(0)
					u.Timeout = 60

					updates, err := tgBot.GetUpdatesChan(u)

					if err != nil {
						log.Fatalf("Error listening to telegram messages: %s", err)
						return
					}

					for update := range updates {
						query := update.CallbackQuery
						if query == nil {
							continue
						}

						if viper.GetInt("tg-admin-id") == query.From.ID {
							log.WithField("user", query.From.UserName).Infof("Skip button pressed in telegram")
							skipChan <- true
						}
					}
				}()
			}

			start = append(start, func(data *models.NotificationData) {
				message := tgbotapi.NewMessage(chatID, generateTelegramMessageText(data, nil))
				message.ParseMode = tgbotapi.ModeMarkdown
				message.ReplyMarkup = messageKeyboard
				send, err := tgBot.Send(message)

				if err != nil {
					log.Errorf("Error sending telegram message: %s", err)
					return
				}

				currentMessage = &send
				lastMessage = time.Now().Unix()
			})

			progressStatus = append(progressStatus, func(data *models.NotificationData) {
				// Rate-limit to 15 messages/min
				if time.Now().Unix()-lastMessage < 4 {
					return
				}

				if currentMessage != nil {
					message := tgbotapi.NewEditMessageText(chatID, currentMessage.MessageID, generateTelegramMessageText(data, nil))
					message.ParseMode = tgbotapi.ModeMarkdown
					message.ReplyMarkup = &messageKeyboard
					_, err := tgBot.Send(message)

					if err != nil {
						log.Errorf("Error editing telegram message: %s", err)
					}

					lastMessage = time.Now().Unix()
				}
			})

			end = append(end, func(data *models.NotificationData, result models.Result) {
				if currentMessage != nil {
					message := tgbotapi.NewEditMessageText(chatID, currentMessage.MessageID, generateTelegramMessageText(data, &result))
					message.ParseMode = tgbotapi.ModeMarkdown
					message.ReplyMarkup = nil
					_, err := tgBot.Send(message)

					if err != nil {
						log.Errorf("Error editing telegram message: %s", err)
					}

					lastMessage = time.Now().Unix()
				}
			})
		}
	})
}

func generateTelegramMessageText(data *models.NotificationData, result *models.Result) string {
	if result != nil && *result == models.ResultError {
		return fmt.Sprintf(
			"*%s*"+
				"\n*Status:* %s",
			data.Filename,
			string(*result),
		)
	}

	diff := (float64(data.CurrentSize) / float64(data.OriginalSize)) * 100

	if result != nil {
		return fmt.Sprintf(
			"*%s*"+
				"\n*Size:* %s --> %s (%.2f%%)"+
				"\n*Status:* %s",
			data.Filename,
			utils.BytesHumanReadable(int64(data.OriginalSize)), utils.BytesHumanReadable(int64(data.CurrentSize)), diff,
			string(*result),
		)
	}

	complete := (float64(data.CurrentFrame) / float64(data.OriginalFrames)) * 100

	expected := "0b"
	eta := time.Duration(0)
	if complete > 0 {
		expected = utils.BytesHumanReadable(int64(float64(data.CurrentSize*100) / complete))
		eta = time.Duration((float64(time.Since(data.Started)) / complete) * (100 - complete))
	}

	return fmt.Sprintf(
		"*%s*"+
			"\n*Size:* %s --> %s (%.2f%%)"+
			"\n*Status:* Transcoding: %.2f%%"+
			"\n*Expected Size:* %s"+
			"\n*ETA:* %s"+
			"\n*FPS:* %.2f",
		data.Filename,
		utils.BytesHumanReadable(int64(data.OriginalSize)), utils.BytesHumanReadable(int64(data.CurrentSize)), diff,
		complete,
		expected,
		eta.Truncate(time.Second),
		data.FPS,
	)
}
