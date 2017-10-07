package loghandler

import (
	"context"
	"fmt"

	"github.com/easymatic/easycontrol/handler"
)

type LogHandler struct {
	handler.BaseHandler
}

func (lh *LogHandler) Start(eventchan chan handler.Event, commandchan chan handler.Command) error {
	lh.EventChan = eventchan
	lh.CommandChanOut = commandchan
	ctx := context.Background()
	lh.Ctx, lh.Cancel = context.WithCancel(ctx)
	fmt.Println("starting log handler")
	for {
		select {
		case event := <-lh.EventChan:
			fmt.Printf("loghander have event: [%s] %s=%s\n", event.Source, event.Tag.Name, event.Tag.Value)
		case <-lh.Ctx.Done():
			fmt.Println("Context canceled")
			return lh.Ctx.Err()
		}
	}
}
