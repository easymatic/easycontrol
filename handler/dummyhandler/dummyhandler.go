package dummyhandler

import (
	"context"
	"fmt"
	"time"

	"github.com/easymatic/easycontrol/handler"
)

type DummyHandler struct {
	handler.Handler
	ctx    context.Context
	cancel context.CancelFunc
}

func (dh *DummyHandler) Start() {
	ctx := context.Background()
	dh.ctx, dh.cancel = context.WithCancel(ctx)
	fmt.Println("starting dummy handler")
	for {
		select {
		case <-time.After(1 * time.Second):
			fmt.Println("Dummyhandler working")
		case <-dh.ctx.Done():
			fmt.Println("Context canceled")
			return
		}
	}
}

func (dh *DummyHandler) Stop() {
	dh.cancel()
	fmt.Println("stopping dummy handler")
}
