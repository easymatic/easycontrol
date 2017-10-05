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

type BaseHandler struct {
	EventChan chan string
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
