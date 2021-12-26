package buttonhandler

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
)

const (
	threshold  = 250
	configPath = "config/buttons.yaml"
)

type tag struct {
	Name  string        `yaml:"name"`
	Event handler.Event `yaml:"event"`
}

type config struct {
	Buttons []struct {
		Name  string        `yaml:"name"`
		Event handler.Event `yaml:"event"`
	} `yaml:"buttons"`
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

type ButtonHandler struct {
	handler.BaseHandler
	config *config
	tags   map[string]*handler.Tag
}

func NewButtonHandler(core handler.CoreHandler) *ButtonHandler {
	rv := &ButtonHandler{}
	rv.Init()
	rv.Name = "buttonhandler"
	rv.CoreHandler = core
	return rv
}

func (hndl *ButtonHandler) Start() error {
	hndl.BaseHandler.Start()
	var err error
	hndl.config, err = getConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get config")
	}
	hndl.EventReader = hndl.CoreHandler.GetEventReader()
	hndl.tags = make(map[string]*handler.Tag, len(hndl.config.Buttons))

	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			for _, button := range hndl.config.Buttons {
				if button.Event.Source == event.Source && button.Event.Tag.Name == event.Tag.Name {
					old, ok := hndl.tags[button.Name]
					if !ok {
						hndl.tags[button.Name] = &event.Tag
						continue
					}
					if old.Value != event.Tag.Value {
						hndl.tags[button.Name] = &event.Tag
						newInt, err := strconv.ParseInt(event.Tag.Value, 10, 16)
						if err != nil {
							log.WithError(err).Errorf("unable to parse value: %s for tag: , skipping", event.Tag.Value, event.Tag.Name)
							continue
						}
						oldInt, err := strconv.ParseInt(old.Value, 10, 16)
						if err != nil {
							log.WithError(err).Errorf("unable to parse value: %s for tag: , skipping", old.Value, event.Tag.Name)
							continue
						}

						if newInt > oldInt {
							for i := oldInt + 1; i < newInt+1; i++ {
								value := strconv.Itoa(int(i & 1))
								hndl.SendEvent(handler.Event{Source: hndl.Name, Tag: handler.Tag{Name: button.Name, Value: value}})
							}
						} else {
							diff := threshold - oldInt + newInt
							for i := oldInt + 1; i < oldInt+diff+1; i++ {
								value := strconv.Itoa(int(i & 1))
								hndl.SendEvent(handler.Event{Source: hndl.Name, Tag: handler.Tag{Name: button.Name, Value: value}})
							}

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
