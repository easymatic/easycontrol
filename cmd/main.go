package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/easymatic/easycontrol/handler/corehandler"
	log "github.com/sirupsen/logrus"
)

func init() {
	// log.SetLevel(log.WarnLevel)
	log.SetLevel(log.DebugLevel)
}

func Start() error {
	core := corehandler.NewCoreHandler()
	if err := core.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	sig := <-sigChan
	log.Infof("Received signal: %v, shutting down gracefully...", sig)

	// Stop all handlers
	core.Stop()
	return nil
}

func main() {
	if err := Start(); err != nil {
		log.WithError(err).Fatal("Failed to start application")
	}
}
