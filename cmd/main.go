package main

import (
	"sync"
	"time"

	"github.com/easymatic/easycontrol/handler/dummyhandler"
	"github.com/easymatic/easycontrol/handler/loghandler"
)

func Start() error {
	// readerhandler.NewArduinoHandler().Start()
	eventchan := make(chan string, 100)
	dummy := &dummyhandler.DummyHandler{}
	log := &loghandler.LogHandler{}
	time.AfterFunc(time.Second*5, dummy.Stop)
	time.AfterFunc(time.Second*5, log.Stop)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Start(eventchan)
	}()
	go func() {
		defer wg.Done()
		dummy.Start(eventchan)
	}()
	wg.Wait()
	return nil
}

func main() {
	Start()
}
