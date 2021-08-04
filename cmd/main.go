package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	telegramwebhook "github.com/runz0rd/telegram-webhook"
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
	for _, w := range c.Webhooks {
		if err := w.ValidateTemplate(); err != nil {
			return err
		}
		th, err := telegramwebhook.NewTelegramHandler(c.BotToken, w.MessageTemplate)
		if err != nil {
			return err
		}
		http.HandleFunc(w.Path, th.Handler)
	}
	log.Printf("serving on :%v", c.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", c.Port), nil)
	if err != nil {
		return err
	}
	return nil
}
