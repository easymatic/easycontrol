package loghandler

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/easymatic/easycontrol/handler"
)

type LogHandler struct {
	handler.BaseHandler
}

func (lh *LogHandler) Start(eventchan chan handler.Event) error {
	lh.EventChan = eventchan
	ctx := context.Background()
	lh.BaseHandler.Ctx, lh.BaseHandler.Cancel = context.WithCancel(ctx)
	fmt.Println("starting log handler")
	for {
		select {
		case event := <-lh.EventChan:
			fmt.Printf("loghander have event: [%s] %s=%s\n", event.Handler, event.SourceId, event.Data)
		case <-lh.BaseHandler.Ctx.Done():
			fmt.Println("Context canceled")
			return lh.BaseHandler.Ctx.Err()
		}
	}
}
