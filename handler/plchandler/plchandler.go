package plchandler

import (
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
	Tags          []Tag
	Values        map[string]string
	ClientHandler *modbus.ASCIIClientHandler
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
func NewPLCHandler() *PLCHandler {
	h := modbus.NewASCIIClientHandler("/dev/ttyPLC")
	h.BaudRate = 9600
	h.DataBits = 7
	h.Parity = "E"
	h.StopBits = 2
	h.SlaveId = 1
	h.Timeout = 2 * time.Second

	rv := &PLCHandler{ClientHandler: h}
	rv.Init()
	rv.Name = "plchandler"
	return rv
}

func (ph *PLCHandler) Start() error {
	ph.BaseHandler.Start()

	ph.Values = make(map[string]string)

	tags := getTagsConfig()
	for _, tag := range tags.Input {
		ph.Tags = append(ph.Tags, tag)
	}
	for _, tag := range tags.Output {
		ph.Tags = append(ph.Tags, tag)
	}
	fmt.Printf("%+v\n", ph.Tags)

	err := ph.ClientHandler.Connect()
	if err != nil {
		return err
	}
	defer ph.ClientHandler.Close()
	//
	client := modbus.NewClient(ph.ClientHandler)
	for {
		select {
		case <-ph.Ctx.Done():
			fmt.Println("Context canceled")
			return ph.Ctx.Err()
		case command := <-ph.CommandChanIn:
			// fmt.Printf("have command: %v\n", command)
			for _, tag := range ph.Tags {
				if tag.Name == command.Name {
					// fmt.Printf("address: %v\n", tag.Address)
					var val uint16 = ON
					if command.Value == "0" {
						val = OFF
					}
					_, err = client.WriteSingleCoil(uint16(tag.Address), val)
					if err != nil {
						fmt.Printf("error: %v\n", err)
					}
				}
			}
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
						ph.SendEvent(handler.Event{Source: "plchandler", Tag: handler.Tag{Name: tag.Name, Value: newValue}})
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
