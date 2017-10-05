package main

import (
	"sync"

	"github.com/easymatic/easycontrol/handler"
	"github.com/easymatic/easycontrol/handler/loghandler"
	"github.com/easymatic/easycontrol/handler/readerhandler"
)

func Start() error {
	// readerhandler.NewArduinoHandler().Start()
	eventchan := make(chan handler.Event, 100)
	//dummy := &dummyhandler.DummyHandler{}
	log := &loghandler.LogHandler{}
	rh := readerhandler.NewReaderHandler()
	//time.AfterFunc(time.Second*5, dummy.Stop)
	//time.AfterFunc(time.Second*5, log.Stop)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Start(eventchan)
	}()
	go func() {
		defer wg.Done()
		rh.Start(eventchan)
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
