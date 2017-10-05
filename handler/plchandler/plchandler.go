package plchandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/easymatic/easycontrol/handler"
	"github.com/goburrow/modbus"
	yaml "gopkg.in/yaml.v2"
)

type PLCHandler struct {
	handler.BaseHandler
}

const ON = 65280 // 0xFF00
const OFF = 0    // 0x0000

type Tags struct {
	Input []struct {
		Name    string `yaml:"name"`
		address int    `yaml:"address"`
	} `yaml:"input"`
	Output []struct {
		Name    string `yaml:"name"`
		address int    `yaml:"address"`
	} `yaml:"output"`
}

func getTagsConfig() Tags {
	var tags Tags
	yamlFile, err := ioutil.ReadFile("config/tags.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &tags)
	if err != nil {
		panic(err)
	}
	return tags
}

func (ph *PLCHandler) Start(eventchan chan string) error {
	ph.EventChan = eventchan
	ctx := context.Background()
	ph.BaseHandler.Ctx, ph.BaseHandler.Cancel = context.WithCancel(ctx)
	fmt.Println("starting plc handler")
	tags := getTagsConfig()
	fmt.Printf("%+v", tags)
	handler := modbus.NewASCIIClientHandler("/dev/ttyPLC")
	handler.BaudRate = 9600
	handler.DataBits = 7
	handler.Parity = "E"
	handler.StopBits = 2
	handler.SlaveId = 1
	handler.Timeout = 2 * time.Second

	err := handler.Connect()
	if err != nil {
		return err
	}
	defer handler.Close()

	client := modbus.NewClient(handler)
	for {
		select {
		case <-ph.BaseHandler.Ctx.Done():
			fmt.Println("Context canceled")
			return ph.BaseHandler.Ctx.Err()
		default:
			results, err := client.ReadCoils(1283, 1)
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
			b := results[0] & 1
			fmt.Printf("readed coil: %v", b)
			// var val uint16 = ON
			// if b != 0 {
			// val = OFF
			// }
			// results, err = client.WriteSingleCoil(1283, val)
			// if err != nil {
			// fmt.Printf("ERROR: %v\n", err)

		}
	}
}
