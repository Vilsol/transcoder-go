package notifications

import (
	"fmt"
	"github.com/Vilsol/transcoder-go/models"
	"github.com/Vilsol/transcoder-go/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

var tgBot *tgbotapi.BotAPI

func init() {
	initialize = append(initialize, func() {
		if viper.GetString("tg-bot-key") != "" && viper.GetInt64("tg-chat-id") != 0 {
			var err error
			tgBot, err = tgbotapi.NewBotAPI(viper.GetString("tg-bot-key"))

			if err != nil {
				log.Fatalf("Error initializing telegram bot: %s", err)
			}

			log.Printf("Telegram connected: %s", tgBot.Self.UserName)

			var currentMessage *tgbotapi.Message

			lastMessage := int64(0)

			start = append(start, func(data *models.NotificationData) {
				message := tgbotapi.NewMessage(viper.GetInt64("tg-chat-id"), generateTelegramMessageText(data, nil))
				message.ParseMode = tgbotapi.ModeMarkdown
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
					message := tgbotapi.NewEditMessageText(viper.GetInt64("tg-chat-id"), currentMessage.MessageID, generateTelegramMessageText(data, nil))
					message.ParseMode = tgbotapi.ModeMarkdown
					_, err := tgBot.Send(message)

					if err != nil {
						log.Errorf("Error editing telegram message: %s", err)
					}

					lastMessage = time.Now().Unix()
				}
			})

			end = append(end, func(data *models.NotificationData, result models.Result) {
				if currentMessage != nil {
					message := tgbotapi.NewEditMessageText(viper.GetInt64("tg-chat-id"), currentMessage.MessageID, generateTelegramMessageText(data, &result))
					message.ParseMode = tgbotapi.ModeMarkdown
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
		eta = time.Duration((float64(time.Now().Sub(data.Started)) / complete) * (100 - complete))
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
