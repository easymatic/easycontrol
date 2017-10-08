package dummyhandler

import (
	"fmt"
	"time"

	"github.com/easymatic/easycontrol/handler"
)

type DummyHandler struct {
	handler.BaseHandler
}

func NewDummyHandler() *DummyHandler {
	rv := &DummyHandler{}
	rv.Init()
	rv.Name = "dummyhandler"
	return rv
}

func (dh *DummyHandler) Start() error {
	dh.BaseHandler.Start()

	val := "0"
	for {
		select {
		case <-time.After(1 * time.Second):
			// dh.SendEvent(handler.Event{Source: "dummyhandler", Tag: handler.Tag{Name: "sometag", Value: "value"}})
			dh.SetTag(handler.Command{Destination: "plchandler", Tag: handler.Tag{Name: "Y3", Value: val}})
			if val == "0" {
				val = "1"
			} else {
				val = "0"
			}
		case <-dh.Ctx.Done():
			fmt.Println("Context canceled")
			return dh.Ctx.Err()
		}
	}
}
