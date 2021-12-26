package dummyhandler

import (
	"time"

	"github.com/easymatic/easycontrol/handler"
	log "github.com/sirupsen/logrus"
)

type DummyHandler struct {
	handler.BaseHandler
}

func NewDummyHandler(core handler.CoreHandler) *DummyHandler {
	rv := &DummyHandler{}
	rv.Init()
	rv.Name = "dummyhandler"
	rv.CoreHandler = core
	return rv
}

func (hndl *DummyHandler) Start() error {
	hndl.BaseHandler.Start()

	for {
		select {
		case <-time.After(1 * time.Second):
			ev := handler.Event{
				Source: "readerhandler",
				Tag: handler.Tag{
					Name:  "Reader0",
					Value: "10636976"}}
			hndl.SendEvent(ev)
		case <-hndl.Ctx.Done():
			log.Info("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
