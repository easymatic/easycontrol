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
	// Tags       map[int]*handler.Tag
	// OutputTags map[int]*handler.Tag
	Tags   []Tag
	Values map[string]string
}

const ON = 65280 // 0xFF00
const OFF = 0    // 0x0000

type Tag struct {
	Name    string `yaml:"name"`
	Address int    `yaml:"address"`
}

type Tags struct {
	Input  []Tag `yaml:"input"`
	Output []Tag `yaml:"output"`
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

func (ph *PLCHandler) Start(eventchan chan handler.Event, commandchan chan handler.Command) error {
	ph.CommandChanOut = commandchan
	ph.Values = make(map[string]string)
	ph.CommandChanIn = make(chan handler.Tag, 100)
	ph.EventChan = eventchan
	ctx := context.Background()
	ph.Ctx, ph.Cancel = context.WithCancel(ctx)
	fmt.Println("starting plc handler")
	tags := getTagsConfig()
	for _, tag := range tags.Input {
		ph.Tags = append(ph.Tags, tag)
	}
	for _, tag := range tags.Output {
		ph.Tags = append(ph.Tags, tag)
	}
	fmt.Printf("%+v\n", ph.Tags)
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
		case command := <-ph.CommandChanIn:
			fmt.Printf("have command: %v\n", command)
		default:
			for _, tag := range ph.Tags {
				results, err := client.ReadCoils(uint16(tag.Address), 1)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
					continue
				}
				b := results[0] & 1
				newValue := strconv.Itoa(int(b))
				value, already := ph.Values[tag.Name]
				if !already {
					fmt.Println("create new one")
					ph.Values[tag.Name] = newValue
				} else {
					if value != newValue {
						ph.Values[tag.Name] = newValue
						ph.SendEvent(handler.Event{Tag: handler.Tag{Name: tag.Name, Value: newValue}})
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
