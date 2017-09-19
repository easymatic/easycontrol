package readerhandler

import (
	"fmt"
	"log"
	"time"

	"github.com/easymatic/easycontrol/handler"
	"github.com/goburrow/modbus"
)

type Tag struct {
	Name  string
	Value string
}

type ArduinoHandler struct {
	handler.Handler
	ClientHandler *modbus.RTUClientHandler
}

func NewArduinoHandler() *ArduinoHandler {
	handler := modbus.NewRTUClientHandler("/dev/ttyMDB")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second
	return &ArduinoHandler{ClientHandler: handler}
}

func (ah *ArduinoHandler) Start() {
	err := ah.ClientHandler.Connect()
	if err != nil {
		log.Fatal(err)
	}

	defer ah.ClientHandler.Close()

	client := modbus.NewClient(ah.ClientHandler)
	for {
		results, _ := client.ReadInputRegisters(0, 28)
		fmt.Printf("Read: %v \n", results)
	}

	//	}
}

func (ah *ArduinoHandler) Stop() {

}
