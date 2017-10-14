package actionhandler

import (
	"fmt"
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
)

const (
	COMMAND_SET   = "set"
	COMMAND_DELAY = "delay"
)

type Config struct {
	Actions []struct {
		Name     string        `yaml:"name"`
		Event    handler.Event `yaml:"event"`
		Commands []struct {
			Name    string          `yaml:"name"`
			Command handler.Command `yaml:"command"`
			Params  interface{}     `yaml:"params"`
		} `yaml:"commands"`
	} `yaml:"actions"`
}

func getActionsConfig() Config {
	var config Config
	yamlFile, err := ioutil.ReadFile("config/actions.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return config
}

type ActionHandler struct {
	handler.BaseHandler
	config Config
}

func NewActionHandler(core handler.CoreHandler) *ActionHandler {
	rv := &ActionHandler{}
	rv.Init()
	rv.Name = "actionhandler"
	rv.config = getActionsConfig()
	rv.CoreHandler = core
	return rv
}

func (hndl *ActionHandler) Start() error {
	hndl.BaseHandler.Start()
	hndl.EventReader = hndl.CoreHandler.GetEventReader()

	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			for _, action := range hndl.config.Actions {
				if action.Event == event {
					for _, command := range action.Commands {
						if command.Name == COMMAND_SET {
							hndl.CoreHandler.RunCommand(command.Command)
						} else if command.Name == COMMAND_DELAY {
							time.Sleep(time.Second * time.Duration(command.Params.(int)))
						}
					}
				}
			}
		case <-hndl.Ctx.Done():
			fmt.Println("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
