package loghandler

import (
	"fmt"

	"github.com/easymatic/easycontrol/handler"
)

type LogHandler struct {
	handler.BaseHandler
}

func NewLogHandler(core handler.CoreHandler) *LogHandler {
	rv := &LogHandler{}
	rv.Init()
	rv.Name = "loghandler"
	rv.CoreHandler = core
	return rv
}

func (hndl *LogHandler) Start() error {
	hndl.BaseHandler.Start()
	hndl.EventReader = hndl.CoreHandler.GetEventReader()

	for {
		select {
		case e := <-hndl.EventReader.Ch:
			event := e.(handler.Event)
			fmt.Printf("loghander have event: [%s] %s=%s\n", event.Source, event.Tag.Name, event.Tag.Value)
		case <-hndl.Ctx.Done():
			fmt.Println("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
