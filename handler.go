package telegramwebhook

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

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
	ChatId          int64  `yaml:"chat_id,omitempty"`
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
	ChatId          int64
	botApi          *tgbotapi.BotAPI
	messageTemplate string
}

func NewTelegramHandler(botToken string, chatId int64, messageTemplate string) (*TelegramHandler, error) {
	ba, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	return &TelegramHandler{chatId, ba, messageTemplate}, nil
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
	data := make(map[string]interface{})
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		return err
	}
	message, err := executeTemplate(th.messageTemplate, data)
	if err != nil {
		return &template.Error{}
	}
	_, err = th.botApi.Send(tgbotapi.NewMessage(th.ChatId, message))
	if err != nil {
		return err
	}
	return nil
}

func executeTemplate(templ string, data map[string]interface{}) (string, error) {
	buf := new(bytes.Buffer)
	if err := template.Must(template.New("").Parse(templ)).Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
