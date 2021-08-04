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
	th, err := telegramwebhook.NewTelegramHandler(c.BotToken, c.ChatId, c.MessageTemplate)
	if err != nil {
		return err
	}
	http.HandleFunc(c.Serve.Path, th.Handler)
	log.Printf("serving on :%v", c.Serve.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", c.Serve.Port), nil)
	if err != nil {
		return err
	}
	return nil
}
