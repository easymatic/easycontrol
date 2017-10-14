package doorhandler

import (
	"fmt"
	"io/ioutil"

	"github.com/easymatic/easycontrol/handler"
	yaml "gopkg.in/yaml.v2"
)

type DoorHandler struct {
	handler.BaseHandler
	config Config
}

type Config struct {
	Doors []struct {
		Name  string        `yaml:"name"`
		Close handler.Event `yaml:"close"`
		Lock  handler.Event `yaml:"lock"`
	} `yaml:"doors"`
}

func getConfig() Config {
	var config Config
	yamlFile, err := ioutil.ReadFile("config/doors.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func NewDoorHandler(core handler.CoreHandler) *DoorHandler {
	rv := &DoorHandler{}
	rv.Init()
	rv.Name = "doorhandler"
	rv.CoreHandler = core
	rv.config = getConfig()
	return rv
}

func (hndl *DoorHandler) Start() error {
	hndl.BaseHandler.Start()
	hndl.EventReader = hndl.CoreHandler.GetEventReader()
	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			value := event.Tag.Value
			event.Tag.Value = ""
			for _, door := range hndl.config.Doors {
				if door.Lock == event {
					ev := handler.Event{
						Source: hndl.GetName(),
						Tag: handler.Tag{
							Name:  fmt.Sprintf("%s.%s", door.Name, "lock"),
							Value: value}}
					hndl.SendEvent(ev)
				} else if door.Close == event {
					ev := handler.Event{
						Source: hndl.GetName(),
						Tag: handler.Tag{
							Name:  fmt.Sprintf("%s.%s", door.Name, "close"),
							Value: value}}
					hndl.SendEvent(ev)
				}
			}
		case <-hndl.Ctx.Done():
			fmt.Println("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
