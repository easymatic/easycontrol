package loghandler

import (
	"context"
	"fmt"

	"github.com/easymatic/easycontrol/handler"
)

type LogHandler struct {
	handler.BaseHandler
}

func (dh *LogHandler) Start(eventchan chan string) {
	dh.EventChan = eventchan
	ctx := context.Background()
	dh.BaseHandler.Ctx, dh.BaseHandler.Cancel = context.WithCancel(ctx)
	fmt.Println("starting log handler")
	for {
		select {
		case event := <-dh.EventChan:
			fmt.Printf("loghander have event: %s\n", event)
		case <-dh.BaseHandler.Ctx.Done():
			fmt.Println("Context canceled")
			return
		}
	}
}
