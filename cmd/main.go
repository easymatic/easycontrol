package main

import (
	"fmt"
	"sync"

	"github.com/easymatic/easycontrol/handler"
	"github.com/easymatic/easycontrol/handler/loghandler"
	"github.com/easymatic/easycontrol/handler/readerhandler"
	"github.com/tjgq/broadcast"
)

func Start() error {
	commandchan := make(chan handler.Command, 100)
	b := broadcast.New(10)
	/*
		dummy := &dummyhandler.DummyHandler{}
		log := &loghandler.LogHandler{}
		rh := &readerhandler.ReaderHandler{}
		plc := &plchandler.PLCHandler{}
	*/
	//dummy := dummyhandler.NewDummyHandler()
	log := loghandler.NewLogHandler()
	rh := readerhandler.NewReaderHandler()
	//	plc := plchandler.NewPLCHandler()

	log.Broadcaster = b
	log.CommandChanOut = commandchan
	//dummy.Broadcaster = b
	//dummy.CommandChanOut = commandchan
	//	plc.Broadcaster = b
	//	plc.CommandChanOut = commandchan
	rh.Broadcaster = b
	rh.CommandChanOut = commandchan
	// time.AfterFunc(time.Second*5, dummy.Stop)
	// time.AfterFunc(time.Second*5, log.Stop)
	// time.AfterFunc(time.Second*5, plc.Stop)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := log.Start(); err != nil {
			fmt.Printf("Error while running log handler: %v\n", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := rh.Start(); err != nil {
			fmt.Printf("Error while running reader handler: %v\n", err)
		}
	}()
	/*
		wg.Add(1)
		go func() {
			defer wg.Done()
			dummy.Start()
		}()
	*/
	/*
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := plc.Start(); err != nil {
				fmt.Printf("Error while running plc handler: %v\n", err)
			}
		}()
	*/
	wg.Wait()
	return nil
}

func main() {
	Start()
}
