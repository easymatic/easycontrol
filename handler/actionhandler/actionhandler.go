package actionhandler

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
)

type Actions struct {
	Actions []struct {
		Name     string        `yaml:"name"`
		Event    handler.Event `yaml:"event"`
		Commands []struct {
			Name    string          `yaml:"name"`
			Command handler.Command `yaml:"command"`
		} `yaml:"commands"`
	} `yaml:"actions"`
}

func getActionsConfig() Actions {
	var actions Actions
	yamlFile, err := ioutil.ReadFile("config/actions.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &actions)
	if err != nil {
		panic(err)
	}
	return actions
}

type ActionHandler struct {
	handler.BaseHandler
	Actions Actions
}

func NewActionHandler() *ActionHandler {
	rv := &ActionHandler{}
	rv.Init()
	rv.Name = "actionhandler"
	rv.Actions = getActionsConfig()
	return rv
}

func (hndl *ActionHandler) Start() error {
	hndl.BaseHandler.Start()
	hndl.EventReader = hndl.Broadcaster.Listen()

	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			for _, action := range hndl.Actions.Actions {
				if action.Event == event {
					for _, command := range action.Commands {
						hndl.SetTag(command.Command)
					}
				}
			}
		case <-hndl.Ctx.Done():
			fmt.Println("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
