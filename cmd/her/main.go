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

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	err := parseFlags()
	if err != nil {
		return err
	}

	err = loadConfig()
	if err != nil {
		return err
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	messagesToBotCh := make(chan her.Message)
	messagesFromBotCh := make(chan her.Message)

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
	var subscriptionConfs []her.SubscriptionConf
	if err := viper.UnmarshalKey("subscriptions", &subscriptionConfs); err != nil {
		return err
	}
	for _, s := range subscriptionConfs {
		if err := m.Subscribe(s); err != nil {
			log.Error(err)
			return err
		}
	}

	var commandConfs []her.CommandConf
	if err := viper.UnmarshalKey("commands", &commandConfs); err != nil {
		return err
	}
	for _, c := range commandConfs {
		if err := b.AddCommand(c); err != nil {
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

func parseFlags() error {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [opts] <config file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "The argument <config file> must be a toml file with a valid configuration\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}
	flag.Parse()

	if *flagVerbose {
		log.SetLevel(log.InfoLevel)
	} else if *flagVeryVerbose {
		log.SetLevel(log.DebugLevel)
	}

	if *flagVersion {
		fmt.Printf("HER - Version %q  https://github.com/tommyblue/her/commit/%s\n", build, build)
		os.Exit(0)
	}

	if len(flag.Args()) < 1 {
		flag.Usage()
		return errors.New("Too few arguments")
	}

	return nil
}

func loadConfig() error {
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return errors.New("Error opening the config file")
	}

	viper.SetConfigType("toml")
	return viper.ReadConfig(f)
}
