package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	err := config()

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
		m, err = mqtt.NewClient(&stopWg, shutdown, messagesFromBotCh, messagesToBotCh)
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
		return err
	}
	for _, subscription := range viper.Get("subscriptions").([]interface{}) {
		topic := subscription.(map[string]interface{})["topic"].(string)
		if err := m.Subscribe(topic); err != nil {
			log.Error(err)
			return err
		}
	}

	if err := b.Connect(); err != nil {
		log.Error(err)
		return err
	}

	log.Info("Waiting to quit..")
	<-quit
	log.Info("Done")

	return nil
}

func config() error {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <config file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "The argument <config file> must be a toml file with a valid configuration\n\n")
	}
	flag.Parse()

	fmt.Println(flag.Arg(0))
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return errors.New("Error opening the config file")
	}

	viper.SetConfigType("toml")
	return viper.ReadConfig(f)
}
