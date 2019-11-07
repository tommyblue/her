package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/tommyblue/her/bot"
	"github.com/tommyblue/her/mqtt"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// log.SetLevel(log.WarnLevel)
}

func main() {
	var err error

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	messagesToBotCh := make(chan mqtt.Message)
	messagesFromBotCh := make(chan mqtt.Message)

	var wg sync.WaitGroup

	var m *mqtt.Client
	wg.Add(1)
	go func() {
		log.Info("Initializing mqtt")
		defer wg.Done()
		m, err = mqtt.NewClient("tcp://192.168.0.100:1883", messagesFromBotCh, messagesToBotCh)
		if err != nil {
			log.Fatal(err)
		}
	}()

	var b *bot.Bot
	wg.Add(1)
	go func() {
		log.Info("Initializing bot")
		defer wg.Done()
		b, err = bot.NewBot(messagesFromBotCh, messagesToBotCh)
		if err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan bool, 1)
	go func() {
		<-shutdown
		log.Info("CTRL+C caught, doing clean shutdown (use CTRL+\\ aka SIGQUIT to abort)")
		b.Stop()
		if err := m.Stop(); err != nil {
			log.Error(err)
		}
		close(messagesToBotCh)
		close(messagesFromBotCh)
		quit <- true
	}()

	wg.Wait()

	if err := m.Connect(); err != nil {
		log.Error(err)
	}
	if err := m.Subscribe("sensor/temperature"); err != nil {
		log.Error(err)
	}
	if err := b.Connect(); err != nil {
		log.Error(err)
	}

	log.Info("Waiting to quit..")
	<-quit
	log.Info("Done")
}
