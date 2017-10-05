package main

import (
	"fmt"
	"sync"

	"github.com/easymatic/easycontrol/handler/dummyhandler"
	"github.com/easymatic/easycontrol/handler/loghandler"
	"github.com/easymatic/easycontrol/handler/plchandler"
)

func Start() error {
	// readerhandler.NewArduinoHandler().Start()
	eventchan := make(chan string, 100)
	dummy := &dummyhandler.DummyHandler{}
	log := &loghandler.LogHandler{}
	plc := &plchandler.PLCHandler{}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := log.Start(eventchan); err != nil {
			fmt.Printf("Error while running log handler: %v\n", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := dummy.Start(eventchan); err != nil {
			fmt.Printf("Error while running dummy handler: %v\n", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := plc.Start(eventchan); err != nil {
			fmt.Printf("Error while running plc handler: %v\n", err)
		}
	}()
	wg.Wait()
	return nil
}

func main() {
	Start()
}
