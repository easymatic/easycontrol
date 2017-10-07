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

func (dh *DummyHandler) Start(eventchan chan handler.Event, commandchan chan handler.Event) error {
	dh.CommandChan = commandchan
	dh.EventChan = eventchan
	ctx := context.Background()
	dh.Ctx, dh.Cancel = context.WithCancel(ctx)
	fmt.Println("starting dummy handler")
	for {
		select {
		case <-time.After(1 * time.Second):
			dh.SendEvent(handler.Event{Handler: "dummyhandler", SourceId: "sometag", Data: "value"})
			dh.SetTag(handler.Event{Handler: "dummyhandler", SourceId: "sometag", Data: "value"})
		case <-dh.Ctx.Done():
			fmt.Println("Context canceled")
			return dh.Ctx.Err()
		}
	}
}
