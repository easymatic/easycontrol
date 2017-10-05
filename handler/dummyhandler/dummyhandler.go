package dummyhandler

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/easymatic/easycontrol/handler"
)

type DummyHandler struct {
	handler.BaseHandler
}

func (dh *DummyHandler) Start(eventchan chan handler.Event) {
	dh.EventChan = eventchan
	ctx := context.Background()
	dh.BaseHandler.Ctx, dh.BaseHandler.Cancel = context.WithCancel(ctx)
	fmt.Println("starting dummy handler")
	for {
		select {
		case <-time.After(1 * time.Second):
			fmt.Println("Dummyhandler working")
			dh.EventChan <- handler.Event{Handler: "dummyhandler", SourceId: "SourceId", Data: "some event from dummy"}
		case <-dh.BaseHandler.Ctx.Done():
			fmt.Println("Context canceled")
			return
		}
	}
}
