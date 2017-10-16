package accesshandler

import (
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/easymatic/easycontrol/handler"
	yaml "gopkg.in/yaml.v2"
)

type AccessHandler struct {
	handler.BaseHandler
	config Config
}

type Config struct {
	Access []struct {
		Name   string        `yaml:"name"`
		Reader handler.Event `yaml:"reader"`
		Cards  []string      `yaml:"cards"`
	} `yaml:"access"`
}

func getConfig() Config {
	var config Config
	yamlFile, err := ioutil.ReadFile("config/access.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func NewAccessHandler(core handler.CoreHandler) *AccessHandler {
	rv := &AccessHandler{}
	rv.Init()
	rv.Name = "accesshandler"
	rv.CoreHandler = core
	rv.config = getConfig()
	fmt.Printf("access: %v", rv.config)
	return rv
}

func (hndl *AccessHandler) Start() error {
	hndl.BaseHandler.Start()
	hndl.EventReader = hndl.CoreHandler.GetEventReader()
	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			value := event.Tag.Value
			event.Tag.Value = ""
			for _, access := range hndl.config.Access {
				if access.Reader == event {
					sort.Strings(access.Cards)
					if i := sort.SearchStrings(access.Cards, value); i < len(access.Cards) {
						ev := handler.Event{
							Source: hndl.GetName(),
							Tag: handler.Tag{
								Name:  access.Name,
								Value: "1"}}
						hndl.SendEvent(ev)
						ev.Tag.Value = "0"
						hndl.SendEvent(ev)
					}

				}
			}

		case <-hndl.Ctx.Done():
			fmt.Println("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
