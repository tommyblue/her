package bot

import (
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tommyblue/her/her"
)

type BotImpl interface {
	Connect() error
	Stop() error
	SendMessage(string) error
}

type Bot struct {
	bot        BotImpl
	stopWg     *sync.WaitGroup
	shutdownCh chan os.Signal
	inCh       <-chan her.Message
	outCh      chan<- her.Message
}

func NewBot(stopWg *sync.WaitGroup, shutdownCh chan os.Signal, outCh, inCh chan her.Message) (*Bot, error) {
	bot := &Bot{
		stopWg:     stopWg,
		shutdownCh: shutdownCh,
		inCh:       inCh,
		outCh:      outCh,
	}

	switch viper.GetString("bot.type") {
	case "telegram":
		telegramBot, err := NewTelegramBot(bot)
		if err != nil {
			return nil, err
		}
		bot.bot = telegramBot
	default:
		log.Panic("Unkown bot")
	}

	return bot, nil
}

func (b *Bot) Connect() error {
	if err := b.bot.Connect(); err != nil {
		fmt.Println("Returning", err)
		return err
	}

	for {
		select {
		case message := <-b.inCh:
			msg := fmt.Sprintf("[%s] %s", message.Topic, message.Message)
			log.Println(msg)
			if err := b.bot.SendMessage(msg); err != nil {
				log.Error(err)
			}
		case <-b.shutdownCh:
			log.Info("Shutting down")
			if err := b.bot.Stop(); err != nil {
				return err
			}
			b.stopWg.Done()
			return nil
		}
	}
}
