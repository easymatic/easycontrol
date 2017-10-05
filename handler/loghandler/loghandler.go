package loghandler

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/easymatic/easycontrol/handler"
)

type LogHandler struct {
	handler.BaseHandler
}

func (dh *LogHandler) Start(eventchan chan handler.Event) {
	dh.EventChan = eventchan
	ctx := context.Background()
	dh.BaseHandler.Ctx, dh.BaseHandler.Cancel = context.WithCancel(ctx)
	fmt.Println("starting log handler")
	for {
		select {
		case event := <-dh.EventChan:
			fmt.Printf("loghander have event: [%s] %s=%s\n", event.Handler, event.SourceId, event.Data)
		case <-dh.BaseHandler.Ctx.Done():
			fmt.Println("Context canceled")
			return
		}
	}
}
