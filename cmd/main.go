package main

import "github.com/easymatic/easycontrol/handler/readerhandler"

func main() {
	h := readerhandler.NewArduinoHandler()
	h.Start()
}
