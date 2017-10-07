package main

import (
	"fmt"
	"sync"

	"github.com/easymatic/easycontrol/handler"
	"github.com/easymatic/easycontrol/handler/dummyhandler"
	"github.com/easymatic/easycontrol/handler/loghandler"
	"github.com/easymatic/easycontrol/handler/plchandler"
)

func Start() error {
	eventchan := make(chan handler.Event, 100)
	commandchan := make(chan handler.Command, 100)
	dummy := &dummyhandler.DummyHandler{}
	log := &loghandler.LogHandler{}
	// rh := &readerhandler.ReaderHandler{}
	plc := &plchandler.PLCHandler{}
	// time.AfterFunc(time.Second*5, dummy.Stop)
	// time.AfterFunc(time.Second*5, log.Stop)
	// time.AfterFunc(time.Second*5, plc.Stop)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := log.Start(eventchan, commandchan); err != nil {
			fmt.Printf("Error while running log handler: %v\n", err)
		}
	}()
	// wg.Add(1)
	// go func() {
	// defer wg.Done()
	// if err := rh.Start(eventchan); err != nil {
	// fmt.Printf("Error while running reader handler: %v\n", err)
	// }
	// }()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := plc.Start(eventchan, commandchan); err != nil {
			fmt.Printf("Error while running plc handler: %v\n", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dummy.Start(eventchan, commandchan)
	}()

	wg.Wait()
	return nil
}

func main() {
	Start()
}
