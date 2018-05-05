package loghandler

import (
	"github.com/easymatic/easycontrol/handler"
	log "github.com/sirupsen/logrus"
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
			log.Infof("loghander have event: [%s] %s=%s\n", event.Source, event.Tag.Name, event.Tag.Value)
		case <-hndl.Ctx.Done():
			log.Infof("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
