package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tommyblue/her/api"
	"github.com/tommyblue/her/bot"
	"github.com/tommyblue/her/her"
	"github.com/tommyblue/her/mqtt"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"
var (
	flagVersion     = flag.Bool("version", false, "Print build version and exit")
	flagVerbose     = flag.Bool("v", false, "verbose logging (info)")
	flagVeryVerbose = flag.Bool("vv", false, "very verbose logging (debug)")
)

func init() {
	log.SetFormatter(&log.TextFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	log.SetLevel(log.WarnLevel)
}

type mainConf struct {
	startWg           sync.WaitGroup
	stopWg            sync.WaitGroup
	messagesToBotCh   chan her.Message
	messagesFromBotCh chan her.Message
	shutdownCh        chan os.Signal
	quitCh            chan bool
	mqtt              *mqtt.Client
	bot               *bot.Bot
}

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	c := mainConf{
		messagesToBotCh:   make(chan her.Message),
		messagesFromBotCh: make(chan her.Message),
		shutdownCh:        make(chan os.Signal, 1),
		quitCh:            make(chan bool, 1),
	}

	if err := c.setup(); err != nil {
		return err
	}

	c.initMQTT()
	c.initBot()
	c.manageShutdown()
	if err := c.runServices(); err != nil {
		return err
	}

	log.Info("Waiting to quit..")
	<-c.quitCh
	log.Info("Done")

	return nil
}

func (c *mainConf) setup() error {
	if err := parseFlags(); err != nil {
		return err
	}

	if err := loadConfig(); err != nil {
		return err
	}

	signal.Notify(c.shutdownCh, os.Interrupt, syscall.SIGTERM)

	return nil
}

func (c *mainConf) initMQTT() {
	c.startWg.Add(1)
	c.stopWg.Add(1)
	go func() {
		defer c.startWg.Done()
		log.Info("Initializing mqtt")
		m, err := mqtt.NewClient(&c.stopWg, c.shutdownCh, c.messagesFromBotCh, c.messagesToBotCh)
		if err != nil {
			log.Fatal(err)
		}
		c.mqtt = m
	}()
}

func (c *mainConf) initBot() {
	c.startWg.Add(1)
	c.stopWg.Add(1)
	go func() {
		defer c.startWg.Done()
		log.Info("Initializing bot")
		b, err := bot.NewBot(&c.stopWg, c.shutdownCh, c.messagesFromBotCh, c.messagesToBotCh)
		if err != nil {
			log.Fatal(err)
		}
		c.bot = b
	}()
}

func (c *mainConf) manageShutdown() {
	go func() {
		<-c.shutdownCh
		log.Info("CTRL+C caught, doing clean shutdown (use CTRL+\\ aka SIGQUIT to abort)")
		close(c.shutdownCh)
		close(c.messagesToBotCh)
		close(c.messagesFromBotCh)
		c.stopWg.Wait()
		c.quitCh <- true
	}()
}

func (c *mainConf) runServices() error {
	c.startWg.Wait()
	if err := c.mqtt.Connect(); err != nil {
		log.Error(err)
		return err
	}
	var subscriptionConfs []her.SubscriptionConf
	if err := viper.UnmarshalKey("subscriptions", &subscriptionConfs); err != nil {
		return err
	}
	for _, s := range subscriptionConfs {
		if err := c.mqtt.Subscribe(s); err != nil {
			log.Error(err)
			return err
		}
	}

	var commandConfs []her.CommandConf
	if err := viper.UnmarshalKey("commands", &commandConfs); err != nil {
		return err
	}
	for _, commandConf := range commandConfs {
		if err := validateCommand(commandConf); err != nil {
			return err
		}
		if err := c.bot.AddCommand(commandConf); err != nil {
			log.Error(err)
			return err
		}
	}

	api.Start(c.messagesFromBotCh)

	if err := c.bot.Connect(); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
