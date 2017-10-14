package plchandler

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"time"

	"github.com/easymatic/easycontrol/handler"
	"github.com/goburrow/modbus"
	yaml "gopkg.in/yaml.v2"
)

type PLCHandler struct {
	handler.BaseHandler
	Tags             map[string]*Tag
	PollingMemBlocks []MemBlock
	ClientHandler    *modbus.ASCIIClientHandler
}

const (
	on  = 65280 // 0xFF00
	off = 0     // 0x0000

	typeInput  = "input"
	typeOutput = "output"

	maxBlockSize = 64
	maxBreakSize = 8
)

type MemBlock struct {
	Address int
	Size    int
	Tags    []*Tag
}

type Tag struct {
	Name    string `yaml:"name"`
	Address int    `yaml:"address"`
	Size    int    `yaml:"-"`
	Type    string `yaml:"-"`
	Value   string `yaml:"-"`
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

func NewPLCHandler(core handler.CoreHandler) *PLCHandler {
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
	rv.CoreHandler = core
	return rv
}

func makePollingMemBlocks(tagsMemBlocks []MemBlock) []MemBlock {
	if len(tagsMemBlocks) == 0 {
		return tagsMemBlocks
	}

	rv := []MemBlock{}
	sortedBlocks := tagsMemBlocks

	sort.Slice(sortedBlocks, func(i, j int) bool {
		return sortedBlocks[i].Address < sortedBlocks[j].Address
	})

	block := MemBlock{Tags: []*Tag{}}
	for _, mb := range sortedBlocks {
		if block.Size == 0 {
			block.Address = mb.Address
			block.Size = mb.Size
			block.Tags = append(block.Tags, mb.Tags[0])
			continue
		}
		newBlockSize := mb.Address + mb.Size - block.Address
		breakSize := mb.Address - block.Address - block.Size
		if newBlockSize > maxBlockSize || breakSize > maxBreakSize {
			rv = append(rv, block)
			block = MemBlock{Address: mb.Address, Size: mb.Size}
			block.Tags = append(block.Tags, mb.Tags[0])
			continue
		}
		block.Size = newBlockSize
		block.Tags = append(block.Tags, mb.Tags[0])
	}
	if block.Size > 0 {
		rv = append(rv, block)
	}

	return rv
}

func (ph *PLCHandler) Start() error {
	ph.BaseHandler.Start()

	ph.Tags = make(map[string]*Tag)

	tags := getTagsConfig()
	tagsCount := len(tags.Input) + len(tags.Output)
	tagsMemBlocks := make([]MemBlock, 0, tagsCount)
	for _, tag := range tags.Input {
		tag.Size = 1
		tag.Type = typeInput
		ph.Tags[tag.Name] = &tag
		mb := MemBlock{Address: tag.Address, Size: tag.Size, Tags: []*Tag{&tag}}
		tagsMemBlocks = append(tagsMemBlocks, mb)
	}
	for _, tag := range tags.Output {
		tag.Size = 1
		tag.Type = typeOutput
		ph.Tags[tag.Name] = &tag
		mb := MemBlock{Address: tag.Address, Size: tag.Size, Tags: []*Tag{&tag}}
		tagsMemBlocks = append(tagsMemBlocks, mb)
	}
	ph.PollingMemBlocks = makePollingMemBlocks(tagsMemBlocks)

	err := ph.ClientHandler.Connect()
	if err != nil {
		return err
	}
	defer ph.ClientHandler.Close()

	client := modbus.NewClient(ph.ClientHandler)
	return ph.loop(&client)
}

func (ph *PLCHandler) loop(client *modbus.Client) error {
	for {
		select {
		case <-ph.Ctx.Done():
			fmt.Println("Context canceled")
			return ph.Ctx.Err()
		case command := <-ph.CommandChanIn:
			//fmt.Printf("have command: %v\n", command)
			tag, ok := ph.Tags[command.Name]
			if ok {
				var val uint16 = on
				if command.Value == "0" {
					val = off
				}
				_, err := (*client).WriteSingleCoil(uint16(tag.Address), val)
				if err != nil {
					fmt.Printf("error: %v\n", err)
				}
			}
		default:
			for _, mb := range ph.PollingMemBlocks {
				results, err := (*client).ReadCoils(uint16(mb.Address), uint16(mb.Size))
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
					continue
				}
				for _, tag := range mb.Tags {
					delta := (*tag).Address - mb.Address
					offs := uint(delta / 8)
					rem := uint(delta % 8)
					mask := 0x80 >> rem
					b := int(results[offs]) & mask
					b >>= (8 - offs)
					newValue := strconv.Itoa(int(b))
					value := tag.Value
					if value != newValue {
						tag.Value = newValue
						ph.SendEvent(handler.Event{Source: "plchandler", Tag: handler.Tag{Name: tag.Name, Value: newValue}})
					}
				}
			}
		}
	}
}
