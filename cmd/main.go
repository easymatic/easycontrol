package main

import "github.com/easymatic/easycontrol/handler/corehandler"

func Start() error {
	core := corehandler.NewCoreHandler()
	core.Start()
	return nil
}

func main() {
	Start()
}
