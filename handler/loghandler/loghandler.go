package loghandler

import (
	"context"
	"fmt"

	"github.com/easymatic/easycontrol/handler"
)

type LogHandler struct {
	handler.BaseHandler
}

func (lh *LogHandler) Start(eventchan chan string) error {
	lh.EventChan = eventchan
	ctx := context.Background()
	lh.BaseHandler.Ctx, lh.BaseHandler.Cancel = context.WithCancel(ctx)
	fmt.Println("starting log handler")
	for {
		select {
		case event := <-lh.EventChan:
			fmt.Printf("loghander have event: %s\n", event)
		case <-lh.BaseHandler.Ctx.Done():
			fmt.Println("Context canceled")
			return lh.BaseHandler.Ctx.Err()
		}
	}
}
