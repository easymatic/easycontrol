package telegramhandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/proxy"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
)

const (
	configPath = "config/telegram.yaml"
)

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

	log.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if cmd := update.Message.CommandWithAt(); cmd != "" {
				log.Infof("command: %s", cmd)
				if cmd == "showtags" {
					tags := make([]string, len(hndl.config.Tags))
					for _, tag := range hndl.config.Tags {
						t, err := hndl.CoreHandler.GetTag(tag.Source, tag.Tag.Name)
						if err != nil {
							log.WithError(err).Error("unable to get current tag value: %v", tag)
							continue
						}
						tags = append(tags, fmt.Sprintf("%s: %s", t.Name, t.Value))
					}
					txt := strings.Join(tags, "\n")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
					bot.Send(msg)
				} else if cmd == "settag" {
					args := update.Message.CommandArguments()
					if args == "" {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "use param - tagname=value")
						bot.Send(msg)
						continue
					}
					argList := strings.Split(args, "=")
					if len(argList) != 2 {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "use param - tagname=value")
						bot.Send(msg)
						continue
					}
					tag := handler.Tag{Name: argList[0], Value: argList[1]}
					cmd := handler.Command{Destination: "plchandler", Tag: tag}
					hndl.CoreHandler.RunCommand(cmd)

				}
			}
		case <-hndl.Ctx.Done():
			log.Info("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
