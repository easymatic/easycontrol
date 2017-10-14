package corehandler

import (
	"fmt"
	"io/ioutil"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"github.com/easymatic/easycontrol/handler"
	"github.com/easymatic/easycontrol/handler/actionhandler"
	"github.com/easymatic/easycontrol/handler/dummyhandler"
	"github.com/easymatic/easycontrol/handler/loghandler"
	"github.com/easymatic/easycontrol/handler/plchandler"
	"github.com/easymatic/easycontrol/handler/readerhandler"
	"github.com/tjgq/broadcast"
)

type Config struct {
	Handlers []struct {
		Name string `yaml:"name"`
		Run  bool   `yaml:"run"`
	}
}

func getHandlersConfig() Config {
	var config Config
	yamlFile, err := ioutil.ReadFile("config/handlers.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return config
}

type CoreHandler struct {
	handler.BaseHandler
	broadcaster *broadcast.Broadcaster
	commandChan chan handler.Command
	handlers    map[string]handler.Handler
	eventReader *broadcast.Listener
	config      Config
}

func NewCoreHandler() *CoreHandler {
	rv := &CoreHandler{}
	rv.Init()
	rv.Name = "corehandler"
	rv.config = getHandlersConfig()
	return rv
}

func (hndl *CoreHandler) GetEventReader() *broadcast.Listener {
	return hndl.broadcaster.Listen()
}

func (hndl *CoreHandler) RunCommand(command handler.Command) {
	for _, h := range hndl.handlers {
		if command.Destination == h.GetName() {
			h.SetTag(command.Tag)
		}
	}
}

func (hndl *CoreHandler) loadHandler() map[string]handler.Handler {
	dummy := dummyhandler.NewDummyHandler(hndl)
	action := actionhandler.NewActionHandler(hndl)
	log := loghandler.NewLogHandler(hndl)
	rh := readerhandler.NewReaderHandler(hndl)
	plc := plchandler.NewPLCHandler(hndl)

	handlers := map[string]handler.Handler{
		dummy.GetName():  dummy,
		rh.GetName():     rh,
		log.GetName():    log,
		plc.GetName():    plc,
		action.GetName(): action,
	}
	return handlers
}

func (hndl *CoreHandler) Start() error {
	hndl.BaseHandler.Start()

	hndl.commandChan = make(chan handler.Command, 100)
	hndl.broadcaster = broadcast.New(10)
	hndl.handlers = hndl.loadHandler()
	var wg sync.WaitGroup
	for _, h := range hndl.config.Handlers {
		fmt.Printf("run: %v\n", h.Run)
		if h.Run {
			if hh, ok := hndl.handlers[h.Name]; ok {
				wg.Add(1)
				go func(hh handler.Handler) {
					defer wg.Done()
					if err := hh.Start(); err != nil {
						fmt.Printf("Error while running handler: %v\n", err)
					}
				}(hh)
			}
		}
	}
	wg.Wait()
	return nil
}

func (hndl *CoreHandler) SendEvent(tag handler.Event) {
	fmt.Printf("sending event in corehandler%v\n", tag)
	hndl.broadcaster.Send(tag)
}
