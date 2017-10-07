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

func (dh *DummyHandler) Start(eventchan chan handler.Event, commandchan chan handler.Command) error {
	dh.CommandChanOut = commandchan
	dh.EventChan = eventchan
	ctx := context.Background()
	dh.Ctx, dh.Cancel = context.WithCancel(ctx)
	fmt.Println("starting dummy handler")
	for {
		select {
		case <-time.After(1 * time.Second):
			dh.SendEvent(handler.Event{Source: "dummyhandler", Tag: handler.Tag{Name: "sometag", Value: "value"}})
			dh.SetTag(handler.Command{Destination: "plchandler", Tag: handler.Tag{Name: "sometag", Value: "value"}})
		case <-dh.Ctx.Done():
			fmt.Println("Context canceled")
			return dh.Ctx.Err()
		}
	}
}
