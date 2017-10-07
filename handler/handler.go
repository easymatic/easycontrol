package handler

import (
	"context"
	"fmt"

	"github.com/tjgq/broadcast"
)

type Handler interface {
	Start() error
	Stop()
	GetTags(key string) string
	SetTag(key string, value string)
}

type Event struct {
	Source string
	Tag    Tag
}

type Command struct {
	Destination string
	Tag         Tag
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

func (bh *BaseHandler) Init() {
	bh.CommandChanIn = make(chan Tag, 100)
}

func (bh *BaseHandler) Start() error {
	fmt.Printf("starting %v handler \n", bh.Name)
	bh.EventReader = bh.Broadcaster.Listen()
	return nil
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
	fmt.Printf("sending event %v\n", tag)
	//bh.EventChan <- tag
	bh.Broadcaster.Send(tag)
	fmt.Println("event sent")
}

type Tag struct {
	Name  string
	Value string
}
