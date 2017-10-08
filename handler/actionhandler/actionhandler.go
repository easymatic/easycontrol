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
	Config Config
}

func NewActionHandler() *ActionHandler {
	rv := &ActionHandler{}
	rv.Init()
	rv.Name = "actionhandler"
	rv.Config = getActionsConfig()
	return rv
}

func (hndl *ActionHandler) Start() error {
	fmt.Printf("actions: %v\n", hndl.Config)
	hndl.BaseHandler.Start()
	hndl.EventReader = hndl.Broadcaster.Listen()

	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			for _, action := range hndl.Config.Actions {
				if action.Event == event {
					for _, command := range action.Commands {
						if command.Name == COMMAND_SET {
							hndl.SetTag(command.Command)
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
