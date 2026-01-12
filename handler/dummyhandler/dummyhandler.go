package dummyhandler

import (
	"sync"
	"time"

	"github.com/easymatic/easycontrol/handler"
	log "github.com/sirupsen/logrus"
)

type DummyHandler struct {
	handler.BaseHandler
	tags map[string]string
	mu   sync.RWMutex
}

func NewDummyHandler(core handler.CoreHandler) *DummyHandler {
	rv := &DummyHandler{}
	rv.Init()
	rv.Name = "dummyhandler"
	rv.CoreHandler = core
	rv.tags = make(map[string]string)
	return rv
}

func (hndl *DummyHandler) GetTag(tagName string) (*handler.Tag, error) {
	hndl.mu.RLock()
	defer hndl.mu.RUnlock()

	value, exists := hndl.tags[tagName]
	if !exists {
		// Return default value "0" for any tag that hasn't been set yet
		value = "0"
	}

	return &handler.Tag{
		Name:  tagName,
		Value: value,
	}, nil
}

func (hndl *DummyHandler) GetTags() []handler.Tag {
	hndl.mu.RLock()
	defer hndl.mu.RUnlock()

	tags := make([]handler.Tag, 0, len(hndl.tags))
	for name, value := range hndl.tags {
		tags = append(tags, handler.Tag{
			Name:  name,
			Value: value,
		})
	}
	return tags
}

func (hndl *DummyHandler) Start() error {
	hndl.BaseHandler.Start()

	// Listen for tag updates from CommandChanIn
	go func() {
		for {
			select {
			case tag := <-hndl.CommandChanIn:
				hndl.mu.Lock()
				oldValue := hndl.tags[tag.Name]
				hndl.tags[tag.Name] = tag.Value
				hndl.mu.Unlock()

				log.Infof("DummyHandler: Set tag %s = %s (was: %s)", tag.Name, tag.Value, oldValue)

				// Send event when tag value changes
				ev := handler.Event{
					Source: hndl.Name,
					Tag:    tag,
				}
				hndl.SendEvent(ev)
			case <-hndl.Ctx.Done():
				return
			}
		}
	}()

	// Keep the original periodic event for readerhandler
	for {
		select {
		case <-time.After(1 * time.Second):
			ev := handler.Event{
				Source: "readerhandler",
				Tag: handler.Tag{
					Name:  "Reader0",
					Value: "10636976"}}
			hndl.SendEvent(ev)
		case <-hndl.Ctx.Done():
			log.Info("Context canceled")
			return hndl.Ctx.Err()
		}
	}
}
