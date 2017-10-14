package handler

import (
	"context"
	"fmt"

	"github.com/tjgq/broadcast"
)

type Handler interface {
	GetName() string
	Start() error
	Stop()
	SetTag(tag Tag)
	GetTag(key string) string
	GetTags() []Tag
}

type CoreHandler interface {
	Handler
	SendEvent(tag Event)
	GetEventReader() *broadcast.Listener
	RunCommand(command Command)
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
	Handler
	Name          string
	CommandChanIn chan Tag
	CoreHandler   CoreHandler
	Ctx           context.Context
	cancel        context.CancelFunc
	EventReader   *broadcast.Listener
}

func (hndl *BaseHandler) Init() {
	hndl.CommandChanIn = make(chan Tag, 100)
	ctx := context.Background()
	hndl.Ctx, hndl.cancel = context.WithCancel(ctx)
}

func (hndl *BaseHandler) Start() error {
	fmt.Printf("starting %v handler \n", hndl.Name)
	return nil
}

func (hndl *BaseHandler) GetCommandChan() chan Tag {
	return hndl.CommandChanIn
}

func (hndl *BaseHandler) GetName() string {
	return hndl.Name
}

func (hndl *BaseHandler) Stop() {
	hndl.cancel()
	hndl.EventReader.Close()
	fmt.Println("stopping dummy handler")
}

func (hndl *BaseHandler) SetTag(tag Tag) {
	// fmt.Printf("setting tag %v\n", command)
	hndl.CommandChanIn <- tag
}

func (hndl *BaseHandler) SendEvent(tag Event) {
	// fmt.Printf("sending event %v\n", tag)
	hndl.CoreHandler.SendEvent(tag)
}

type Tag struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
