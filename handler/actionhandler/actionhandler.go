package actionhandler

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
)

const (
	commandSet    = "set"
	commandDelay  = "delay"
	commandInvert = "invert"
	configPath    = "config/actions.yaml"
)

type config struct {
	Actions []struct {
		Name       string          `yaml:"name"`
		Event      handler.Event   `yaml:"event"`
		Conditions []handler.Event `yaml:"conditions"`
		Commands   []struct {
			Name    string          `yaml:"name"`
			Command handler.Command `yaml:"command"`
			Params  interface{}     `yaml:"params"`
		} `yaml:"commands"`
	} `yaml:"actions"`
}

func getActionsConfig() (*config, error) {
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

type ActionHandler struct {
	handler.BaseHandler
	config *config
}

func NewActionHandler(core handler.CoreHandler) *ActionHandler {
	rv := &ActionHandler{}
	rv.Init()
	rv.Name = "actionhandler"
	rv.CoreHandler = core
	return rv
}

func (hndl *ActionHandler) checkConditions(conditions []handler.Event) bool {
	for _, c := range conditions {
		t, err := hndl.CoreHandler.GetTag(c.Source, c.Tag.Name)
		if err != nil {
			log.WithError(err).Error("unable to get current tag value: %v", c)
			continue
		}
		if t.Value != c.Tag.Value {
			return false
		}
	}
	return true
}

func (hndl *ActionHandler) Start() error {
	hndl.BaseHandler.Start()
	var err error
	hndl.config, err = getActionsConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get config")
	}
	hndl.EventReader = hndl.CoreHandler.GetEventReader()

	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			for _, action := range hndl.config.Actions {
				if action.Event == event && hndl.checkConditions(action.Conditions) {
					for _, command := range action.Commands {
						if command.Name == commandSet {
							log.Infof("run command set: %v", command.Command)
							hndl.CoreHandler.RunCommand(command.Command)
						} else if command.Name == commandDelay {
							log.Infof("run command delay: %v", command.Params)
							time.Sleep(time.Second * time.Duration(command.Params.(int)))
						} else if command.Name == commandInvert {
							log.Infof("run command invert: %v", command.Command)
							t, err := hndl.CoreHandler.GetTag(command.Command.Destination, command.Command.Tag.Name)
							if err != nil {
								log.WithError(err).Error("unable to get current tag value: %v", command)
								continue
							}
							command.Command.Tag.Value = "1"
							if t.Value == "1" {
								command.Command.Tag.Value = "0"
							}
							hndl.CoreHandler.RunCommand(command.Command)
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
