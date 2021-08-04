package telegramwebhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v3"
)

type Serve struct {
	Path string
	Port int
}

type Config struct {
	Serve           Serve  `yaml:"serve,omitempty"`
	BotToken        string `yaml:"bot_token,omitempty"`
	MessageTemplate string `yaml:"message_template,omitempty"`
}

func ReadConfig(path string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type TelegramHandler struct {
	botApi          *tgbotapi.BotAPI
	messageTemplate string
}

func NewTelegramHandler(botToken string, messageTemplate string) (*TelegramHandler, error) {
	ba, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	return &TelegramHandler{ba, messageTemplate}, nil
}

func (th TelegramHandler) Handler(w http.ResponseWriter, req *http.Request) {
	err := th.handle(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
	}
	return
}

func (th TelegramHandler) handle(req *http.Request) error {
	uriSlice := strings.Split(req.URL.Path, "/")
	if len(uriSlice) < 2 {
		return fmt.Errorf("need to specify telegram chat id at the end of the request uri (/webhook/12345)")
	}
	chatId, err := strconv.ParseInt(uriSlice[len(uriSlice)-1], 10, 64)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		return err
	}
	message, err := executeTemplate(th.messageTemplate, data)
	if err != nil {
		return &template.Error{}
	}
	if message == "" {
		log.Print("message empty, nothing sent")
		return nil
	}
	_, err = th.botApi.Send(tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           chatId,
			ReplyToMessageID: 0,
		},
		Text:                  message,
		ParseMode:             tgbotapi.ModeMarkdown,
		DisableWebPagePreview: false,
	})
	if err != nil {
		return err
	}
	log.Printf("successfully sent %q to %q", message, chatId)
	return nil
}

func executeTemplate(templ string, data map[string]interface{}) (string, error) {
	buf := new(bytes.Buffer)
	if err := template.Must(template.New("").Parse(templ)).Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
