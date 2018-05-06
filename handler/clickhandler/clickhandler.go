package clickhandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
)

const (
	configPath = "config/clicks.yaml"
)

type tag struct {
	Name  string        `yaml:"name"`
	Event handler.Event `yaml:"event"`
}

type config struct {
	Clicks []struct {
		Name      string        `yaml:"name"`
		Event     handler.Event `yaml:"event"`
		OnRelease bool          `yaml:"on_release"`
		Count     int           `yaml:"count"`
		Timeout   int           `yaml:"timeout"`
	} `yaml:"clicks"`
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
	for i, click := range c.Clicks {
		if click.Count == 0 {
			c.Clicks[i].Count = 1
		}
	}
	return c, nil
}

type ClickHandler struct {
	handler.BaseHandler
	config *config
	tags   map[string]context.CancelFunc
}

func NewClickHandler(core handler.CoreHandler) *ClickHandler {
	rv := &ClickHandler{}
	rv.Init()
	rv.Name = "clickhandler"
	rv.CoreHandler = core
	return rv
}

func (hndl *ClickHandler) Start() error {
	hndl.BaseHandler.Start()
	var err error
	hndl.config, err = getConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get config")
	}
	hndl.EventReader = hndl.CoreHandler.GetEventReader()
	hndl.tags = make(map[string]context.CancelFunc, len(hndl.config.Clicks))

	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			for _, click := range hndl.config.Clicks {
				if click.Event.Source == event.Source && click.Event.Tag.Name == event.Tag.Name {
					if t, ok := hndl.tags[event.Tag.Name]; ok {
						t()
					}
					if click.Timeout > 0 && ((click.OnRelease && event.Tag.Value == "0") || (!click.OnRelease && event.Tag.Value == "1")) {
						ctx, cancel := context.WithTimeout(context.Background(), time.Duration(click.Timeout)*time.Millisecond)
						hndl.tags[event.Tag.Name] = cancel
						go func() {
							log.Info("start timeout func")
							select {
							case <-ctx.Done():
								if ctx.Err() == context.DeadlineExceeded {
									hndl.SendEvent(handler.Event{Source: hndl.Name, Tag: handler.Tag{Name: click.Name, Value: "1"}})
								}
								delete(hndl.tags, event.Tag.Name)
							}
						}()
					}
					log.Infof("click %v", event)
				}
			}
		case <-hndl.Ctx.Done():
			log.Info("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
