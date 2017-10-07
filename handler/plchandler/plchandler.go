package plchandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/easymatic/easycontrol/handler"
	"github.com/goburrow/modbus"
	yaml "gopkg.in/yaml.v2"
)

type PLCHandler struct {
	handler.BaseHandler
	Tags map[int]*handler.Tag
}

const ON = 65280 // 0xFF00
const OFF = 0    // 0x0000

type Tags struct {
	Input []struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
	} `yaml:"input"`
	Output []struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
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

func (ph *PLCHandler) Start(eventchan chan handler.Event, commandchan chan handler.Event) error {
	ph.Tags = make(map[int]*handler.Tag)
	ph.CommandChan = commandchan
	ph.EventChan = eventchan
	ctx := context.Background()
	ph.Ctx, ph.Cancel = context.WithCancel(ctx)
	fmt.Println("starting plc handler")
	tags := getTagsConfig()
	fmt.Printf("%+v\n", tags)
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
	//
	client := modbus.NewClient(h)
	for {
		select {
		case <-ph.Ctx.Done():
			fmt.Println("Context canceled")
			return ph.Ctx.Err()
		case command := <-ph.CommandChan:
			fmt.Printf("have command: %v\n", command)
		default:
			for _, tag := range tags.Input {
				results, err := client.ReadCoils(uint16(tag.Address), 1)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
				}
				if len(results) > 0 {
					b := results[0] & 1
					// fmt.Printf("read %s: %v\n", tag.Name, b)
					t, already := ph.Tags[tag.Address]
					if !already {
						fmt.Println("create new one")
						t = &handler.Tag{Name: tag.Name}
						ph.Tags[tag.Address] = t
					}
					val := strconv.Itoa(int(b))
					// fmt.Printf("%s and %s\n", t.Value, val)
					if t.Value != val {
						t.Value = val
						ph.SendEvent(handler.Event{SourceId: t.Name, Handler: "plchandler", Data: t.Value})
					}
				}

			}
			// time.Sleep(time.Second)
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
