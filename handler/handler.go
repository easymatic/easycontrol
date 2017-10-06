package handler

import (
	"fmt"

	"golang.org/x/net/context"
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
	EventChan chan Event
	Handler
	Ctx    context.Context
	Cancel context.CancelFunc
}

func (bh *BaseHandler) Stop() {
	bh.Cancel()
	fmt.Println("stopping dummy handler")
}

type Tag struct {
	Name  string
	Value string
}
