package loghandler

import (
	"fmt"

	"github.com/easymatic/easycontrol/handler"
)

type LogHandler struct {
	handler.BaseHandler
}

func NewLogHandler() *LogHandler {
	rv := &LogHandler{}
	rv.Init()
	rv.Name = "loghandler"
	return rv
}

func (lh *LogHandler) Start() error {
	lh.BaseHandler.Start()
	lh.EventReader = lh.Broadcaster.Listen()

	for {
		select {
		case e := <-lh.EventReader.Ch:
			event := e.(handler.Event)
			fmt.Printf("loghander have event: [%s] %s=%s\n", event.Source, event.Tag.Name, event.Tag.Value)
		case <-lh.Ctx.Done():
			fmt.Println("Context canceled")
			return lh.Ctx.Err()
		}
	}
}
