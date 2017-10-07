package handler

import (
	"context"
	"fmt"
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
	EventChan      chan Event
	CommandChanIn  chan Tag
	CommandChanOut chan Command
	Handler
	Ctx    context.Context
	Cancel context.CancelFunc
}

func (bh *BaseHandler) Stop() {
	bh.Cancel()
	fmt.Println("stopping dummy handler")
}

func (bh *BaseHandler) SetTag(command Command) {
	fmt.Printf("setting tag %v\n", command)
	bh.CommandChanOut <- command
}

func (bh *BaseHandler) SendEvent(tag Event) {
	fmt.Printf("sending event %v\n", tag)
	bh.EventChan <- tag
}

type Tag struct {
	Name  string
	Value string
}
