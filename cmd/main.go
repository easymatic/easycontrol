package main

import (
	"time"

	"github.com/easymatic/easycontrol/handler/dummyhandler"
)

func Start() error {
	// readerhandler.NewArduinoHandler().Start()
	dummy := &dummyhandler.DummyHandler{}
	time.AfterFunc(time.Second*5, dummy.Stop)
	dummy.Start()
	// time.Sleep(time.Second * 8)
	return nil
}

func main() {
	Start()
}
