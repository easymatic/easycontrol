package main

import (
	"github.com/easymatic/easycontrol/handler/corehandler"
	log "github.com/sirupsen/logrus"
)

func init() {
	// log.SetLevel(log.WarnLevel)
	log.SetLevel(log.DebugLevel)
}

func Start() error {
	core := corehandler.NewCoreHandler()
	core.Start()
	return nil
}

func main() {
	Start()
}
