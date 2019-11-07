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

	var startWg sync.WaitGroup
	var stopWg sync.WaitGroup

	var m *mqtt.Client
	startWg.Add(1)
	stopWg.Add(1)
	go func() {
		defer startWg.Done()
		log.Info("Initializing mqtt")
		m, err = mqtt.NewClient(&stopWg, shutdown, "tcp://test.mosquitto.org:1883", messagesFromBotCh, messagesToBotCh)
		if err != nil {
			log.Fatal(err)
		}
	}()

	var b *bot.Bot
	startWg.Add(1)
	stopWg.Add(1)
	go func() {
		defer startWg.Done()
		log.Info("Initializing bot")
		b, err = bot.NewBot(&stopWg, shutdown, messagesFromBotCh, messagesToBotCh)
		if err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan bool, 1)
	go func() {
		<-shutdown
		log.Info("CTRL+C caught, doing clean shutdown (use CTRL+\\ aka SIGQUIT to abort)")
		close(shutdown)
		close(messagesToBotCh)
		close(messagesFromBotCh)
		stopWg.Wait()
		quit <- true
	}()

	startWg.Wait()
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
