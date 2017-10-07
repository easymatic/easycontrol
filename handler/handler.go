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
	Handler  string
	SourceId string
	Data     string
}

type BaseHandler struct {
	EventChan   chan Event
	CommandChan chan Event
	Handler
	Ctx    context.Context
	Cancel context.CancelFunc
}

func (bh *BaseHandler) Stop() {
	bh.Cancel()
	fmt.Println("stopping dummy handler")
}

func (bh *BaseHandler) SetTag(tag Event) {
	fmt.Printf("setting tag %v\n", tag)
	bh.CommandChan <- tag
}

func (bh *BaseHandler) SendEvent(tag Event) {
	fmt.Printf("sending event %v\n", tag)
	bh.EventChan <- tag
}

type Tag struct {
	Name  string
	Value string
}
