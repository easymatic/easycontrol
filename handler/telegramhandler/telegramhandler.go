package telegramhandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/proxy"

	// tgbotapi "gopkg.in/telegram-bot-api.v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
	msgs   map[int64]int
	bot    *tgbotapi.BotAPI
}

func NewTelegramHandler(core handler.CoreHandler) *TelegramHandler {
	rv := &TelegramHandler{}
	rv.msgs = map[int64]int{}
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
			log.WithError(err).Errorf("unable to get current tag value: %v", tag)
			continue
		}
		tags[i] = handler.Event{Source: tag.Source, Tag: handler.Tag{Name: t.Name, Value: t.Value}}
	}
	buttons := [][]tgbotapi.InlineKeyboardButton{}
	row := []tgbotapi.InlineKeyboardButton{}
	size := 2
	for _, tag := range tags {
		state := "⚪"
		if tag.Tag.Value == "0" {
			tag.Tag.Value = "1"
			state = "⚫"
		} else {
			tag.Tag.Value = "0"
		}
		cmd := fmt.Sprintf("%s:%s:%s", tag.Source, tag.Tag.Name, tag.Tag.Value)
		status := fmt.Sprintf("%s: %s", tag.Tag.Name, state)
		btn := tgbotapi.NewInlineKeyboardButtonData(status, cmd)
		row = append(row, btn)
		if len(row) == size {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(row...))
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	mrk := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return mrk
}

func (hndl *TelegramHandler) handleEvents() error {
	hndl.EventReader = hndl.CoreHandler.GetEventReader()

	for {
		select {
		case e := <-hndl.EventReader.Ch:
			if hndl.bot == nil {
				continue
			}
			event := e.(handler.Event)
			for _, tag := range hndl.config.Tags {
				if event.Source == tag.Source && event.Tag.Name == tag.Tag.Name {
					fmt.Println("here")
					for chatID, msgID := range hndl.msgs {
						mrk := hndl.getInlineKeyboard()
						edit := tgbotapi.NewEditMessageReplyMarkup(chatID, msgID, mrk)
						if _, err := hndl.bot.Send(edit); err != nil {
							log.WithError(err).Error("unable to send keyboard update")
						}
					}
				}
			}
		case <-hndl.Ctx.Done():
			log.Info("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}

func (hndl *TelegramHandler) Start() error {
	hndl.BaseHandler.Start()
	var err error
	hndl.config, err = getConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get config")
	}
	go hndl.handleEvents()

	if hndl.config.Proxy != nil && hndl.config.Proxy.Address != "" {
		dialer, err := proxy.SOCKS5("tcp", hndl.config.Proxy.Address, &proxy.Auth{User: hndl.config.Proxy.UserName, Password: hndl.config.Proxy.Password}, proxy.Direct)
		if err != nil {
			return errors.Wrap(err, "unable to connect to proxy")
		}

		httpTransport := &http.Transport{}
		httpClient := &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial
		hndl.bot, err = tgbotapi.NewBotAPIWithClient(hndl.config.Token, tgbotapi.APIEndpoint, httpClient)
		if err != nil {
			return errors.Wrap(err, "unable to connect to telegram api")
		}
	} else {
		hndl.bot, err = tgbotapi.NewBotAPI(hndl.config.Token)
		if err != nil {
			return errors.Wrap(err, "unable to connect to telegram api")
		}
	}
	hndl.bot.Debug = true

	log.Infof("new version Authorized on account %s", hndl.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	updates := hndl.bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			go func() {
				log.Infof("------------------start!!! processing-------------------------------")
				if update.CallbackQuery != nil {
					if !checkAccess(update.CallbackQuery.From.UserName) {
						config := tgbotapi.CallbackConfig{}
						config.CallbackQueryID = update.CallbackQuery.ID
						config.Text = "access denied"
						// if _, err := hndl.bot.AnswerCallbackQuery(config); err != nil {
						// log.WithError(err).Error("unable to send access denied")
						// }
						log.Infof("------------------end processing-------------------------------")
						return
					}
					ss := strings.Split(update.CallbackQuery.Data, ":")
					cmd := &handler.Command{
						Destination: ss[0],
						Tag: handler.Tag{
							Name:  ss[1],
							Value: ss[2],
						},
					}
					// if err := json.Unmarshal([]byte(update.CallbackQuery.Data), cmd); err != nil {
					// log.WithError(err).Errorf("unable unmarshal json: %s", update.CallbackQuery.Data)
					// }
					hndl.CoreHandler.RunCommand(*cmd)
					config := tgbotapi.CallbackConfig{}
					config.CallbackQueryID = update.CallbackQuery.ID
					config.Text = "Done"
					// time.Sleep(time.Duration(500) * time.Millisecond)
					mrk := hndl.getInlineKeyboard()
					edit := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, mrk)
					if _, err := hndl.bot.Send(edit); err != nil {
						log.WithError(err).Error("unable to send keyboard update")
					}
					// if _, err := hndl.bot.AnswerCallbackQuery(config); err != nil {
					// log.WithError(err).Error("unable to send done")
					// }
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
						if _, err := hndl.bot.Send(msg); err != nil {
							log.WithError(err).Error("unable to send starting")
						}
					case "show":
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "tags:")
						mrk := hndl.getInlineKeyboard()
						msg.ReplyMarkup = mrk
						m, err := hndl.bot.Send(msg)
						if err != nil {
							log.WithError(err).Error("unable to send tags with keyboard")
							log.Infof("------------------end processing-------------------------------")
							return
						}
						hndl.msgs[update.Message.Chat.ID] = m.MessageID
					case "settag":
						if !checkAccess(update.Message.From.UserName) {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "access denied")
							hndl.bot.Send(msg)
							log.Infof("------------------end processing-------------------------------")
							return
						}
						args := update.Message.CommandArguments()
						if args == "" {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "use param - tagname=value")
							hndl.bot.Send(msg)
							log.Infof("------------------end processing-------------------------------")
							return
						}
						argList := strings.Split(args, "=")
						if len(argList) != 2 {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "use param - tagname=value")
							hndl.bot.Send(msg)
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
