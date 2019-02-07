package main

import (
	"time"

	"github.com/easymatic/easycontrol/handler/corehandler"
	log "github.com/sirupsen/logrus"
)

func init() {
	// log.SetLevel(log.WarnLevel)
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	log.SetLevel(log.DebugLevel)
	// log.SetFormatter(&log.TextFormatter{
	// DisableColors: false,
	// FullTimestamp: true,
	// })
}

func Start() error {
	core := corehandler.NewCoreHandler()
	core.Start()
	return nil
}

func main() {
	Start()
}
