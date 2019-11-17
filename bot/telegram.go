package bot

import (
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tommyblue/her/her"
)

type TelegramBot struct {
	api       *tgbotapi.BotAPI
	token     string
	channelId int64
	bot       *Bot
}

func NewTelegramBot(bot *Bot) (*TelegramBot, error) {
	token := viper.GetString("bot.token")
	if token == "" {
		return nil, errors.New("Missing bot token")
	}

	channelId := viper.GetInt64("bot.channel_id")
	if channelId == 0 {
		return nil, errors.New("Missing channel id")
	}

	return &TelegramBot{
		token:     token,
		channelId: channelId,
		bot:       bot,
	}, nil
}

func (t *TelegramBot) Connect() error {
	botApi, err := tgbotapi.NewBotAPI(t.token)
	if err != nil {
		return err
	}
	t.api = botApi

	t.api.Debug = false

	log.Info("Authorized on account ", t.api.Self.UserName)

	return t.SendMessage("Hi! I've been just started")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := t.api.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	go func() {
		for update := range updates {
			t.messageReceived(update)
		}
	}()

	return nil
}

func (t *TelegramBot) Stop() error {
	log.Info("Stopping telegram bot")
	return t.SendMessage("Bye bye")
}

func (t *TelegramBot) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(t.channelId, message)
	_, err := t.api.Send(msg)
	return err
}

func (t *TelegramBot) messageReceived(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	log.Info(fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

	if update.Message.IsCommand() {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "help":
			msg.Text = "type /sayhi or /status."
		case "sayhi":
			msg.Text = "Hi :)"
		case "status":
			msg.Text = "I'm ok."
		case "on":
			msg.Text = t.newCommandReceived("on", update.Message.CommandArguments())
		case "off":
			msg.Text = t.newCommandReceived("off", update.Message.CommandArguments())
		default:
			msg.Text = "I don't know that command"
		}
		if _, err := t.api.Send(msg); err != nil {
			log.Error(err)
		}
	}
}

func (t *TelegramBot) newCommandReceived(command, arguments string) string {
	switch command {
	case "on":
		t.bot.outCh <- her.Message{Topic: "homeassistant/switch1", Message: []byte("ON")}
		return "Switched on"
	case "off":
		t.bot.outCh <- her.Message{Topic: "homeassistant/switch1", Message: []byte("OFF")}
		return "Switched off"
	default:
		return "Wrong command"
	}
}
