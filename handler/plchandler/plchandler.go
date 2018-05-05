package plchandler

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"time"

	"github.com/easymatic/easycontrol/handler"
	"github.com/goburrow/modbus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type PLCHandler struct {
	handler.BaseHandler
	tags             map[string]*Tag
	pollingMemBlocks []MemBlock
	clientHandler    *modbus.ASCIIClientHandler
}

const (
	on  = 65280 // 0xFF00
	off = 0     // 0x0000

	typeInput  = "input"
	typeOutput = "output"

	maxBlockSize = 64
	maxBreakSize = 8
	configPath   = "config/tags.yaml"
)

type MemBlock struct {
	address uint16
	size    uint16
	tags    []*Tag
}

type Tag struct {
	Name    string `yaml:"name"`
	Address uint16 `yaml:"address"`
	Size    uint16 `yaml:"-"`
	Type    string `yaml:"-"`
	Value   string `yaml:"-"`
}

type Config struct {
	Input  []Tag  `yaml:"input"`
	Output []Tag  `yaml:"output"`
	Device string `yaml:"device"`
}

func getConfig() (*Config, error) {
	config := &Config{}
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to open config: %s", configPath))
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to parse config: %s", configPath))
	}
	return config, nil
}

func NewPLCHandler(core handler.CoreHandler) *PLCHandler {
	rv := &PLCHandler{}
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

	sort.Slice(tagsMemBlocks, func(i, j int) bool {
		return tagsMemBlocks[i].address < tagsMemBlocks[j].address
	})

	block := MemBlock{tags: []*Tag{}}
	for _, mb := range tagsMemBlocks {
		if block.size == 0 {
			block.address = mb.address
			block.size = mb.size
			block.tags = append(block.tags, mb.tags[0])
			continue
		}
		newBlockSize := mb.address + mb.size - block.address
		breakSize := mb.address - block.address - block.size
		if newBlockSize > maxBlockSize || breakSize > maxBreakSize {
			rv = append(rv, block)
			block = MemBlock{address: mb.address, size: mb.size}
			block.tags = append(block.tags, mb.tags[0])
			continue
		}
		block.size = newBlockSize
		block.tags = append(block.tags, mb.tags[0])
	}
	if block.size > 0 {
		rv = append(rv, block)
	}

	return rv
}

func (ph *PLCHandler) Start() error {
	ph.BaseHandler.Start()

	ph.tags = make(map[string]*Tag)

	config, err := getConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get config")
	}
	tagsCount := len(config.Input) + len(config.Output)
	tagsMemBlocks := make([]MemBlock, 0, tagsCount)
	for _, tag := range config.Input {
		t := Tag{Size: 1, Type: typeInput, Address: tag.Address, Name: tag.Name}
		ph.tags[tag.Name] = &t
		mb := MemBlock{address: t.Address, size: t.Size, tags: []*Tag{&t}}
		tagsMemBlocks = append(tagsMemBlocks, mb)
	}
	for _, tag := range config.Output {
		t := Tag{Size: 1, Type: typeOutput, Address: tag.Address, Name: tag.Name}
		ph.tags[tag.Name] = &t
		mb := MemBlock{address: t.Address, size: t.Size, tags: []*Tag{&t}}
		tagsMemBlocks = append(tagsMemBlocks, mb)
	}
	ph.pollingMemBlocks = makePollingMemBlocks(tagsMemBlocks)
	for _, m := range ph.pollingMemBlocks {
		log.Infof("address: %v, size: %v", m.address, m.size)
		for _, t := range m.tags {
			log.Infof("%v | ", t)
		}
	}

	h := modbus.NewASCIIClientHandler(config.Device)
	h.BaudRate = 9600
	h.DataBits = 7
	h.Parity = "E"
	h.StopBits = 2
	h.SlaveId = 1
	h.Timeout = 2 * time.Second

	ph.clientHandler = h
	err = ph.clientHandler.Connect()
	if err != nil {
		return errors.Wrap(err, "unable to open serial device")
	}
	defer ph.clientHandler.Close()

	client := modbus.NewClient(ph.clientHandler)
	return ph.loop(client)
}

func (ph *PLCHandler) loop(client modbus.Client) error {
	for {
		select {
		case <-ph.Ctx.Done():
			log.Info("Context canceled")
			return ph.Ctx.Err()
		case command := <-ph.CommandChanIn:
			tag, ok := ph.tags[command.Name]
			if ok {
				var val uint16 = on
				if command.Value == "0" {
					val = off
				}
				if _, err := client.WriteSingleCoil(tag.Address, val); err != nil {
					log.WithError(err).Errorf("unable to write coil to addr %v: %v", tag.Address, val)
				}
			}
		default:
			for _, mb := range ph.pollingMemBlocks {
				r, err := client.ReadHoldingRegisters(3597, 1)
				if err != nil {
					log.WithError(err).Errorf("unable to read coil: %v", 3597)
					continue
				}
				log.Infof("len: %d\n", len(r))
				log.Infof("register: %d, %d\n", r[0], r[1])
				results, err := client.ReadCoils(mb.address, mb.size)
				if err != nil {
					log.WithError(err).Errorf("unable to read coils: %v, with size: %v", mb.address, mb.size)
					continue
				}
				for _, tag := range mb.tags {
					delta := tag.Address - mb.address
					offs := delta / 8
					rem := delta % 8
					mask := 0x01 << rem
					b := int(results[offs]) & mask
					b >>= rem
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
