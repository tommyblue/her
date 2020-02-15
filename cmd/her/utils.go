package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tommyblue/her/her"
)

func validateCommand(command her.CommandConf) error {
	if command.Help == "" {
		return fmt.Errorf("Command /%s is missing the help", command.Command)
	}

	if command.Command == "" {
		return fmt.Errorf("Command is empty")
	}

	if command.Topic == "" {
		return fmt.Errorf("Command /%s is missing the topic", command.Command)
	}

	if command.Message == "" {
		return fmt.Errorf("Command /%s is missing the message", command.Command)
	}

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

func loadConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return errors.New("Error opening the config file")
	}

	viper.SetConfigType("toml")
	return viper.ReadConfig(f)
}
