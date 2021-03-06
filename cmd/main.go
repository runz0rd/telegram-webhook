package main

import (
	"flag"
	"fmt"
	"net/http"

	telegramwebhook "github.com/runz0rd/telegram-webhook"
	log "github.com/sirupsen/logrus"
)

func main() {
	var flagConfig string
	flag.StringVar(&flagConfig, "config", "config.yaml", "config file location")
	flag.Parse()

	if err := run(flagConfig); err != nil {
		log.Fatal(err)
	}
}

func run(config string) error {
	c, err := telegramwebhook.ReadConfig(config)
	if err != nil {
		return err
	}
	if c.Debug {
		log.SetLevel(log.DebugLevel)
	}
	for _, w := range c.Webhooks {
		if err := w.ValidateTemplate(); err != nil {
			return err
		}
		th, err := telegramwebhook.NewTelegramHandler(c.BotToken, w.MessageTemplate, w.DeduplicateRangeSecond)
		if err != nil {
			return err
		}
		http.HandleFunc(w.GetPath(), th.Handler)
		log.Debugf("handling path %v", w.GetPath())
	}
	log.Printf("serving on :%v", c.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", c.Port), nil)
	if err != nil {
		return err
	}
	return nil
}
