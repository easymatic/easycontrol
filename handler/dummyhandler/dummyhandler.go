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

func (dh *DummyHandler) Start(eventchan chan string) error {
	dh.EventChan = eventchan
	ctx := context.Background()
	dh.BaseHandler.Ctx, dh.BaseHandler.Cancel = context.WithCancel(ctx)
	fmt.Println("starting dummy handler")
	for {
		select {
		case <-time.After(1 * time.Second):
			fmt.Println("Dummyhandler working")
			dh.EventChan <- "some event from dummy"
		case <-dh.BaseHandler.Ctx.Done():
			fmt.Println("Context canceled")
			return dh.BaseHandler.Ctx.Err()
		}
	}
}
