package main

import (
	"fmt"
	"sync"

	"github.com/easymatic/easycontrol/handler"
	"github.com/easymatic/easycontrol/handler/loghandler"
	"github.com/easymatic/easycontrol/handler/plchandler"
	"github.com/easymatic/easycontrol/handler/readerhandler"
)

func Start() error {
	eventchan := make(chan handler.Event, 100)
	//dummy := &dummyhandler.DummyHandler{}
	log := &loghandler.LogHandler{}
	rh := &readerhandler.ReaderHandler{}
	//time.AfterFunc(time.Second*5, dummy.Stop)
	//time.AfterFunc(time.Second*5, log.Stop)
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
		rh.Start(eventchan)
		if err := rh.Start(eventchan); err != nil {
			fmt.Printf("Error while running reader handler: %v\n", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := plc.Start(eventchan); err != nil {
			fmt.Printf("Error while running plc handler: %v\n", err)
		}
	}()

	/*
		go func() {
			defer wg.Done()
			dummy.Start(eventchan)
		}()
	*/
	wg.Wait()
	return nil
}

func main() {
	Start()
}
