package telegramhandler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/proxy"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
)

const (
	configPath = "config/telegram.yaml"
)

var users = []string{"aborilov", "agniya9"}

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
	yamlFile, err := os.ReadFile(configPath)
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
	tags := make([]handler.Event, 0, len(hndl.config.Tags))
	for _, tag := range hndl.config.Tags {
		t, err := hndl.CoreHandler.GetTag(tag.Source, tag.Tag.Name)
		if err != nil {
			// Silently skip tags that can't be retrieved
			continue
		}
		tags = append(tags, handler.Event{Source: tag.Source, Tag: handler.Tag{Name: t.Name, Value: t.Value}})
	}

	// Create buttons in a 2-column grid layout
	var buttons [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(tags); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// First button in row
		if i < len(tags) {
			tag := tags[i]
			emoji := "💡"
			newValue := "1"
			if tag.Tag.Value == "0" {
				newValue = "1"
				emoji = "⚫"
			} else {
				newValue = "0"
				emoji = "💡"
			}
			// Use compact format: "source:tagName:value" (much shorter than JSON)
			callbackData := fmt.Sprintf("%s:%s:%s", tag.Source, tag.Tag.Name, newValue)
			status := fmt.Sprintf("%s %s", emoji, tag.Tag.Name)
			btn := tgbotapi.NewInlineKeyboardButtonData(status, callbackData)
			row = append(row, btn)
		}

		// Second button in row (if exists)
		if i+1 < len(tags) {
			tag := tags[i+1]
			emoji := "💡"
			newValue := "1"
			if tag.Tag.Value == "0" {
				newValue = "1"
				emoji = "⚫"
			} else {
				newValue = "0"
				emoji = "💡"
			}
			// Use compact format: "source:tagName:value" (much shorter than JSON)
			callbackData := fmt.Sprintf("%s:%s:%s", tag.Source, tag.Tag.Name, newValue)
			status := fmt.Sprintf("%s %s", emoji, tag.Tag.Name)
			btn := tgbotapi.NewInlineKeyboardButtonData(status, callbackData)
			row = append(row, btn)
		}

		if len(row) > 0 {
			buttons = append(buttons, row)
		}
	}

	// Add refresh button at the bottom
	refreshBtn := tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "refresh")
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{refreshBtn})

	mrk := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return mrk
}

func (hndl *TelegramHandler) getStatusMessage() string {
	var statusLines []string
	statusLines = append(statusLines, "🏠 *Home Control Status*\n")

	onCount := 0
	offCount := 0

	for _, tag := range hndl.config.Tags {
		t, err := hndl.CoreHandler.GetTag(tag.Source, tag.Tag.Name)
		if err != nil {
			// Silently skip tags that can't be retrieved
			continue
		}
		emoji := "⚫"
		if t.Value == "1" {
			emoji = "💡"
			onCount++
		} else {
			offCount++
		}
		statusLines = append(statusLines, fmt.Sprintf("%s %s", emoji, t.Name))
	}

	statusLines = append(statusLines, fmt.Sprintf("\n📊 *Summary:* %d ON | %d OFF", onCount, offCount))

	return strings.Join(statusLines, "\n")
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
		bot, err = tgbotapi.NewBotAPIWithClient(hndl.config.Token, tgbotapi.APIEndpoint, httpClient)
		if err != nil {
			return errors.Wrap(err, "unable to connect to telegram api")
		}
	} else {
		bot, err = tgbotapi.NewBotAPI(hndl.config.Token)
		if err != nil {
			return errors.Wrap(err, "unable to connect to telegram api")
		}
	}
	bot.Debug = false

	log.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			go func() {
				if update.CallbackQuery != nil {
					if update.CallbackQuery.Data == "refresh" {
						mrk := hndl.getInlineKeyboard()
						statusMsg := hndl.getStatusMessage()
						edit := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, statusMsg)
						edit.ParseMode = "Markdown"
						edit.ReplyMarkup = &mrk
						if _, err := bot.Send(edit); err != nil {
							log.WithError(err).Error("unable to send refresh update")
						}
						config := tgbotapi.NewCallback(update.CallbackQuery.ID, "✅ Refreshed")
						if _, err := bot.Request(config); err != nil {
							log.WithError(err).Error("unable to send callback answer")
						}
						return
					}
					if !checkAccess(update.CallbackQuery.From.UserName) {
						config := tgbotapi.NewCallback(update.CallbackQuery.ID, "❌ Access denied")
						config.ShowAlert = true
						if _, err := bot.Request(config); err != nil {
							log.WithError(err).Error("unable to send access denied")
						}
						return
					}
					// Parse compact format: "source:tagName:value"
					parts := strings.Split(update.CallbackQuery.Data, ":")
					if len(parts) != 3 {
						log.Errorf("Invalid callback data format: %s", update.CallbackQuery.Data)
						config := tgbotapi.NewCallback(update.CallbackQuery.ID, "❌ Error processing command")
						config.ShowAlert = true
						bot.Request(config)
						return
					}

					cmd := handler.Command{
						Destination: parts[0],
						Tag: handler.Tag{
							Name:  parts[1],
							Value: parts[2],
						},
					}

					// Execute command
					hndl.CoreHandler.RunCommand(cmd)

					// Show feedback
					config := tgbotapi.NewCallback(update.CallbackQuery.ID, "✅ Done")
					if _, err := bot.Request(config); err != nil {
						log.WithError(err).Error("unable to send callback answer")
					}

					// Update keyboard after a short delay
					time.Sleep(300 * time.Millisecond)
					mrk := hndl.getInlineKeyboard()
					statusMsg := hndl.getStatusMessage()
					edit := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, statusMsg)
					edit.ParseMode = "Markdown"
					edit.ReplyMarkup = &mrk
					if _, err := bot.Send(edit); err != nil {
						log.WithError(err).Error("unable to send keyboard update")
					}
					return
				}
				if update.Message == nil {
					return
				}

				cmd := update.Message.CommandWithAt()

				if cmd != "" {
					switch cmd {
					case "start":
						welcomeMsg := `🏠 *Welcome to Home Control Bot!*

Use the buttons below to control your home devices.

*Available commands:*
/show - Show all controls
/status - Show current status
/help - Show help message

Tap the button below to get started! 👇`
						btn := tgbotapi.NewKeyboardButton("🏠 Show Controls")
						kb := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn})
						kb.ResizeKeyboard = true
						kb.OneTimeKeyboard = true
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMsg)
						msg.ParseMode = "Markdown"
						msg.ReplyMarkup = kb
						if _, err := bot.Send(msg); err != nil {
							log.WithError(err).Error("unable to send welcome message")
						}
					case "show", "controls":
						statusMsg := hndl.getStatusMessage()
						mrk := hndl.getInlineKeyboard()
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, statusMsg)
						msg.ParseMode = "Markdown"
						msg.ReplyMarkup = mrk
						_, err := bot.Send(msg)
						if err != nil {
							log.WithError(err).Error("unable to send controls")
						}
					case "status":
						statusMsg := hndl.getStatusMessage()
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, statusMsg)
						msg.ParseMode = "Markdown"
						if _, err := bot.Send(msg); err != nil {
							log.WithError(err).Error("unable to send status")
						}
					case "help":
						helpMsg := `📖 *Help - Home Control Bot*

*Commands:*
/show or /controls - Show all device controls
/status - Show current device status
/help - Show this help message

*Usage:*
• Tap any button to toggle a device ON/OFF
• Tap 🔄 Refresh to update the status
• Buttons show current state: 💡 = ON, ⚫ = OFF

*Access:*
Only authorized users can control devices.`
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpMsg)
						msg.ParseMode = "Markdown"
						if _, err := bot.Send(msg); err != nil {
							log.WithError(err).Error("unable to send help")
						}
					case "settag":
						if !checkAccess(update.Message.From.UserName) {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Access denied")
							bot.Send(msg)
							return
						}
						args := update.Message.CommandArguments()
						if args == "" {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Usage: /settag handler:tagname=value\nExample: /settag dummyhandler:BoilerRoom=1")
							bot.Send(msg)
							return
						}
						argList := strings.Split(args, "=")
						if len(argList) != 2 {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Invalid format. Use: /settag handler:tagname=value\nExample: /settag dummyhandler:BoilerRoom=1")
							bot.Send(msg)
							return
						}

						tagSpec := strings.TrimSpace(argList[0])
						tagValue := strings.TrimSpace(argList[1])

						// Handler must be specified in format "handler:tagname"
						if !strings.Contains(tagSpec, ":") {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Handler must be specified. Use: handler:tagname=value\nExample: /settag dummyhandler:BoilerRoom=1")
							bot.Send(msg)
							return
						}

						parts := strings.SplitN(tagSpec, ":", 2)
						if len(parts) != 2 {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Invalid format. Use: handler:tagname=value\nExample: /settag dummyhandler:BoilerRoom=1")
							bot.Send(msg)
							return
						}

						destination := strings.TrimSpace(parts[0])
						tagName := strings.TrimSpace(parts[1])

						tag := handler.Tag{Name: tagName, Value: tagValue}
						cmd := handler.Command{Destination: destination, Tag: tag}
						hndl.CoreHandler.RunCommand(cmd)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Set %s = %s (via %s)", tag.Name, tag.Value, destination))
						bot.Send(msg)
					}
				} else {
					// Handle button text clicks (non-command messages)
					if update.Message.Text == "🏠 Show Controls" {
						statusMsg := hndl.getStatusMessage()
						mrk := hndl.getInlineKeyboard()
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, statusMsg)
						msg.ParseMode = "Markdown"
						msg.ReplyMarkup = mrk
						if _, err := bot.Send(msg); err != nil {
							log.WithError(err).Error("unable to send controls from button")
						}
					}
				}
			}()
		case <-hndl.Ctx.Done():
			log.Info("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
