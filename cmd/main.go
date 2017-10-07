package main

import (
	"fmt"
	"sync"

	"github.com/easymatic/easycontrol/handler"
	"github.com/easymatic/easycontrol/handler/actionhandler"
	"github.com/easymatic/easycontrol/handler/loghandler"
	"github.com/easymatic/easycontrol/handler/plchandler"
	"github.com/tjgq/broadcast"
)

func Start() error {
	commandchan := make(chan handler.Command, 100)
	b := broadcast.New(10)
	// dummy := dummyhandler.NewDummyHandler()
	action := actionhandler.NewActionHandler()
	log := loghandler.NewLogHandler()
	// rh := readerhandler.NewReaderHandler()
	plc := plchandler.NewPLCHandler()

	handlers := []handler.Handler{}
	// handlers = append(handlers, dummy)
	// handlers = append(handlers, rh)
	handlers = append(handlers, log)
	handlers = append(handlers, plc)
	handlers = append(handlers, action)
	// handlers = append(handlers, dummy)
	var wg sync.WaitGroup
	for _, h := range handlers {
		h.SetBroadcaster(b)
		h.SetCommandChan(commandchan)
		wg.Add(1)
		go func(h handler.Handler) {
			defer wg.Done()
			if err := h.Start(); err != nil {
				fmt.Printf("Error while running log handler: %v\n", err)
			}
		}(h)
	}
	// wg.Add(1)
	// go func() {
	// defer wg.Done()
	// if err := log.Start(); err != nil {
	// fmt.Printf("Error while running log handler: %v\n", err)
	// }
	// }()

	for {
		select {
		case command := <-commandchan:
			for _, h := range handlers {
				if command.Destination == h.GetName() {
					h.GetCommandChan() <- command.Tag
				}
			}
		}
	}
}

func main() {
	Start()
}
