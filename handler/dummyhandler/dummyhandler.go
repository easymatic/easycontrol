package dummyhandler

import (
	"context"
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

	ctx := context.Background()
	dh.Ctx, dh.Cancel = context.WithCancel(ctx)

	for {
		select {
		case <-time.After(1 * time.Second):
			dh.SendEvent(handler.Event{Source: dh.Name, Tag: handler.Tag{Name: "sometag", Value: "value"}})
			dh.SetTag(handler.Command{Destination: "plchandler", Tag: handler.Tag{Name: "sometag", Value: "value"}})
		case <-dh.Ctx.Done():
			fmt.Println("Context canceled")
			return dh.Ctx.Err()
		}
	}
}
