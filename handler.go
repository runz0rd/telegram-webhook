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
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Webhook struct {
	Path            string `yaml:"path,omitempty"`
	MessageTemplate string `yaml:"message_template,omitempty"`
}

type Config struct {
	Webhooks []Webhook `yaml:"webhooks,omitempty"`
	BotToken string    `yaml:"bot_token,omitempty"`
	Port     int       `yaml:"port,omitempty"`
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
		err = errors.Wrapf(err, "[%v]", req.URL.Path)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
	}
	return
}

func (th TelegramHandler) handle(req *http.Request) error {
	if req.Method != "POST" {
		return fmt.Errorf("you need to use POST")
	}
	uriSlice := strings.Split(req.URL.Path, "/")
	if len(uriSlice) < 2 || uriSlice[len(uriSlice)-1] == "" {
		return fmt.Errorf("you need to specify the telegram chat id at the end of the request uri (/webhook/12345)")
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
		return fmt.Errorf("message empty, nothing sent")
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
	log.Printf("[%v]: successfully sent %q to %d", req.URL.Path, message, chatId)
	return nil
}

func executeTemplate(templ string, data map[string]interface{}) (string, error) {
	buf := new(bytes.Buffer)
	if err := template.Must(template.New("").Parse(templ)).Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
