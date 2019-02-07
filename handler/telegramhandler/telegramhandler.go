package telegramhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/proxy"

	"gopkg.in/telegram-bot-api.v4"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
)

const (
	configPath = "config/telegram.yaml"
)

var users = []string{"aborilov", "agniya9"}

type tag struct {
	Name  string        `yaml:"name"`
	Event handler.Event `yaml:"event"`
}

type config struct {
	Proxy *struct {
		Address  string `yaml:"address"`
		UserName string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"proxy"`
	Token string           `yaml:"token"`
	Tags  []*handler.Event `yaml:"tags"`
}

func checkAccess(user string) bool {
	for _, u := range users {
		if u == user {
			return true
		}
	}
	return false
}

func getConfig() (*config, error) {
	c := &config{}
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to open config: %s", configPath))
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to parse config: %s", configPath))
	}
	return c, nil
}

type TelegramHandler struct {
	handler.BaseHandler
	config *config
	tags   map[string]context.CancelFunc
}

func NewTelegramHandler(core handler.CoreHandler) *TelegramHandler {
	rv := &TelegramHandler{}
	rv.Init()
	rv.Name = "telegramhandler"
	rv.CoreHandler = core
	return rv
}

func (hndl *TelegramHandler) getInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	tags := make([]handler.Event, len(hndl.config.Tags))
	for i, tag := range hndl.config.Tags {
		t, err := hndl.CoreHandler.GetTag(tag.Source, tag.Tag.Name)
		if err != nil {
			log.WithError(err).Error("unable to get current tag value: %v", tag)
			continue
		}
		tags[i] = handler.Event{Source: tag.Source, Tag: handler.Tag{Name: t.Name, Value: t.Value}}
	}
	buttons := make([][]tgbotapi.InlineKeyboardButton, len(tags)+1)
	for i, tag := range tags {
		state := "on"
		if tag.Tag.Value == "0" {
			tag.Tag.Value = "1"
			state = "off"
		} else {
			tag.Tag.Value = "0"
		}
		cmd := handler.Command{Destination: tag.Source, Tag: tag.Tag}
		data, err := json.Marshal(cmd)
		if err != nil {
			log.WithError(err).Error("unable to marshal json: %v", cmd)
			continue
		}
		status := fmt.Sprintf("%s: %s", tag.Tag.Name, state)
		btn := tgbotapi.NewInlineKeyboardButtonData(status, string(data))
		row := tgbotapi.NewInlineKeyboardRow(btn)
		buttons[i] = row
	}
	btn := tgbotapi.NewInlineKeyboardButtonData("refresh", "refresh")
	row := tgbotapi.NewInlineKeyboardRow(btn)
	buttons[len(hndl.config.Tags)] = row
	mrk := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return mrk
}

func (hndl *TelegramHandler) Start() error {
	hndl.BaseHandler.Start()
	var err error
	hndl.config, err = getConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get config")
	}
	// hndl.EventReader = hndl.CoreHandler.GetEventReader()

	var bot *tgbotapi.BotAPI
	if hndl.config.Proxy != nil {
		dialer, err := proxy.SOCKS5("tcp", hndl.config.Proxy.Address, &proxy.Auth{User: hndl.config.Proxy.UserName, Password: hndl.config.Proxy.Password}, proxy.Direct)
		if err != nil {
			return errors.Wrap(err, "unable to connect to proxy")
		}

		httpTransport := &http.Transport{}
		httpClient := &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial
		bot, err = tgbotapi.NewBotAPIWithClient(hndl.config.Token, httpClient)
		if err != nil {
			return errors.Wrap(err, "unable to connect to telegram api")
		}
	} else {
		bot, err = tgbotapi.NewBotAPI(hndl.config.Token)
		if err != nil {
			return errors.Wrap(err, "unable to connect to telegram api")
		}
	}
	bot.Debug = true

	log.Infof("new version Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	updates, err := bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			go func() {
				log.Infof("------------------start!!! processing-------------------------------")
				if update.CallbackQuery != nil {
					if update.CallbackQuery.Data == "refresh" {
						log.Info("get new keyboard")
						mrk := hndl.getInlineKeyboard()
						log.Info("have new keyboard")
						edit := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, mrk)
						log.Info("have edit msg")
						if _, err := bot.Send(edit); err != nil {
							log.WithError(err).Error("unable to send refresh update")
						}
						log.Info("msg sent")
						config := tgbotapi.CallbackConfig{}
						config.CallbackQueryID = update.CallbackQuery.ID
						config.Text = "Done"
						if _, err := bot.AnswerCallbackQuery(config); err != nil {
							log.WithError(err).Error("unable to send done")
						}
						log.Infof("------------------end processing-------------------------------")
						return
					}
					if !checkAccess(update.CallbackQuery.From.UserName) {
						config := tgbotapi.CallbackConfig{}
						config.CallbackQueryID = update.CallbackQuery.ID
						config.Text = "access denied"
						if _, err := bot.AnswerCallbackQuery(config); err != nil {
							log.WithError(err).Error("unable to send access denied")
						}
						log.Infof("------------------end processing-------------------------------")
						return
					}
					cmd := &handler.Command{}
					if err := json.Unmarshal([]byte(update.CallbackQuery.Data), cmd); err != nil {
						log.WithError(err).Error("unable unmarshal json: %s", update.CallbackQuery.Data)
					}
					hndl.CoreHandler.RunCommand(*cmd)
					config := tgbotapi.CallbackConfig{}
					config.CallbackQueryID = update.CallbackQuery.ID
					config.Text = "Done"
					time.Sleep(time.Duration(500) * time.Millisecond)
					mrk := hndl.getInlineKeyboard()
					edit := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, mrk)
					if _, err := bot.Send(edit); err != nil {
						log.WithError(err).Error("unable to send keyboard update")
					}
					if _, err := bot.AnswerCallbackQuery(config); err != nil {
						log.WithError(err).Error("unable to send done")
					}
					log.Infof("------------------end processing-------------------------------")
					return
				}
				if update.Message == nil {
					log.Infof("------------------end processing-------------------------------")
					return
				}
				log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
				if cmd := update.Message.CommandWithAt(); cmd != "" {
					log.Infof("command: %s", cmd)
					switch cmd {
					case "start":
						btn := tgbotapi.NewKeyboardButton("/show")
						kb := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn})
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "starting...")
						msg.ReplyMarkup = kb
						if _, err := bot.Send(msg); err != nil {
							log.WithError(err).Error("unable to send starting")
						}
					case "show":
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "tags:")
						mrk := hndl.getInlineKeyboard()
						msg.ReplyMarkup = mrk
						_, err := bot.Send(msg)
						if err != nil {
							log.WithError(err).Error("unable to send tags with keyboard")
							log.Infof("------------------end processing-------------------------------")
							return
						}
					case "settag":
						if !checkAccess(update.Message.From.UserName) {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "access denied")
							bot.Send(msg)
							log.Infof("------------------end processing-------------------------------")
							return
						}
						args := update.Message.CommandArguments()
						if args == "" {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "use param - tagname=value")
							bot.Send(msg)
							log.Infof("------------------end processing-------------------------------")
							return
						}
						argList := strings.Split(args, "=")
						if len(argList) != 2 {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "use param - tagname=value")
							bot.Send(msg)
							log.Infof("------------------end processing-------------------------------")
							return
						}
						tag := handler.Tag{Name: argList[0], Value: argList[1]}
						cmd := handler.Command{Destination: "plchandler", Tag: tag}
						hndl.CoreHandler.RunCommand(cmd)
					}
				}
				log.Infof("------------------end processing-------------------------------")
			}()
		case <-hndl.Ctx.Done():
			log.Info("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
