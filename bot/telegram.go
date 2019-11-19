package bot

import (
	"errors"
	"fmt"
	"strings"

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
	commands  map[string]her.CommandConf
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
		commands:  make(map[string]her.CommandConf),
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

	if err := t.SendMessage("Hi! I've been just started"); err != nil {
		log.Error(err)
	}

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

func (t *TelegramBot) AddCommand(c her.CommandConf) error {
	_, ok := t.commands[c.Command]
	if ok {
		return fmt.Errorf("Command %s already exists", c.Command)
	}
	t.commands[c.Command] = c

	return nil
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
			msg.Text = t.printHelp()
		default:
			msg.Text = t.checkCommands(update.Message.Command(), update.Message.CommandArguments())
		}
		if _, err := t.api.Send(msg); err != nil {
			log.Error(err)
		}
	}
}

func (t *TelegramBot) printHelp() string {
	var b strings.Builder
	b.WriteString("Available commands:\n\n")
	b.WriteString("/help - Get this help\n")
	for command, conf := range t.commands {
		b.WriteString(fmt.Sprintf("/%s - %s\n", command, conf.Help))
	}
	return b.String()
}

func (t *TelegramBot) checkCommands(command, args string) string {
	cmd, ok := t.commands[command]
	if !ok {
		log.Error("Unknown command: ", command)
		return "I don't know that command"
	}
	t.bot.outCh <- her.Message{Topic: cmd.Topic, Message: []byte(cmd.Message)}
	return cmd.FeedbackMsg
}
