package plchandler

import (
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/net/context"

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

func (ph *PLCHandler) Start(eventchan chan handler.Event) error {
	ph.EventChan = eventchan
	ctx := context.Background()
	ph.BaseHandler.Ctx, ph.BaseHandler.Cancel = context.WithCancel(ctx)
	fmt.Println("starting plc handler")
	tags := getTagsConfig()
	fmt.Printf("%+v", tags)
	h := modbus.NewASCIIClientHandler("/dev/ttyPLC")
	h.BaudRate = 9600
	h.DataBits = 7
	h.Parity = "E"
	h.StopBits = 2
	h.SlaveId = 1
	h.Timeout = 2 * time.Second

	err := h.Connect()
	if err != nil {
		return err
	}
	defer h.Close()

	client := modbus.NewClient(h)
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
			if len(results) > 0 {
				b := results[0] & 1
				fmt.Printf("readed coil: %v\n", b)
			}
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
