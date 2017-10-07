package handler

import (
	"context"
	"fmt"

	"github.com/tjgq/broadcast"
)

type Handler interface {
	Stop()
	Start() error
	GetTags(key string) string
	SetTag(command Command)
	GetCommandChan() chan Tag
	GetName() string
	SetBroadcaster(broadcaster *broadcast.Broadcaster)
	SetCommandChan(commandchan chan Command)
}

type Event struct {
	Source string `yaml:"source"`
	Tag    Tag    `yaml:"tag"`
}

type Command struct {
	Destination string `yaml:"destination"`
	Tag         Tag    `yaml:"tag"`
}

type BaseHandler struct {
	Name           string
	CommandChanIn  chan Tag
	CommandChanOut chan Command
	Handler
	Ctx         context.Context
	Cancel      context.CancelFunc
	EventReader *broadcast.Listener
	Broadcaster *broadcast.Broadcaster
}

func (bh *BaseHandler) SetCommandChan(commandchan chan Command) {
	bh.CommandChanOut = commandchan
}

func (bh *BaseHandler) SetBroadcaster(broadcaster *broadcast.Broadcaster) {
	bh.Broadcaster = broadcaster
}

func (bh *BaseHandler) Init() {
	bh.CommandChanIn = make(chan Tag, 100)
	ctx := context.Background()
	bh.Ctx, bh.Cancel = context.WithCancel(ctx)
}

func (bh *BaseHandler) Start() error {
	fmt.Printf("starting %v handler \n", bh.Name)
	return nil
}

func (bh *BaseHandler) GetCommandChan() chan Tag {
	return bh.CommandChanIn
}

func (bh *BaseHandler) GetName() string {
	return bh.Name
}

func (bh *BaseHandler) Stop() {
	bh.Cancel()
	bh.EventReader.Close()
	fmt.Println("stopping dummy handler")
}

func (bh *BaseHandler) SetTag(command Command) {
	fmt.Printf("setting tag %v\n", command)
	bh.CommandChanOut <- command
}

func (bh *BaseHandler) SendEvent(tag Event) {
	// fmt.Printf("sending event %v\n", tag)
	bh.Broadcaster.Send(tag)
}

type Tag struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
