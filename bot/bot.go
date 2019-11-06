package bot

import (
	"errors"
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/tommyblue/her/mqtt"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	token  string
	stopCh chan bool
	inCh   chan mqtt.Message
}

func NewBot(inCh chan mqtt.Message) (*Bot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, errors.New("Missing bot token")
	}
	bot := &Bot{
		token:  token,
		stopCh: make(chan bool),
		inCh:   inCh,
	}

	return bot, nil
}

func (b *Bot) Connect() error {
	botApi, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		return err
	}
	b.api = botApi

	b.api.Debug = false

	log.Info("Authorized on account ", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for {
		select {
		case message := <-b.inCh:
			b.sendMessage(fmt.Sprintf("[%s] %s", message.Topic, message.Message))
		case <-b.stopCh:
			return nil
		case update := <-updates:
			b.messageReceived(update)
		}
	}
}

func (b *Bot) Stop() {
	log.Info("Stopping bot")
	b.stopCh <- true
}

func (b *Bot) messageReceived(update tgbotapi.Update) {
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
			msg.Text = "Switch on"
		case "off":
			msg.Text = "Switch off"
		default:
			msg.Text = "I don't know that command"
		}
		if _, err := b.api.Send(msg); err != nil {
			log.Error(err)
		}
	}
}

func (b *Bot) sendMessage(message string) {
	msg := tgbotapi.NewMessage(158066827, message)
	if _, err := b.api.Send(msg); err != nil {
		log.Panic(err)
	}
}
