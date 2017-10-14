package readerhandler

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/easymatic/easycontrol/handler"
	"github.com/goburrow/modbus"
)

const READER_START_ADDRESS = 16
const READER_BLOCK_SIZE = 2
const READER_COUNT = 2

type Reader struct {
	EventId  int
	CardCode int
}

type ReaderHandler struct {
	handler.BaseHandler

	ClientHandler *modbus.RTUClientHandler
	Readers       [READER_COUNT]Reader
}

func NewReaderHandler(core handler.CoreHandler) *ReaderHandler {
	h := modbus.NewRTUClientHandler("/dev/ttyMDB")
	h.BaudRate = 9600
	h.DataBits = 8
	h.Parity = "N"
	h.StopBits = 1
	h.SlaveId = 1
	h.Timeout = 5 * time.Second

	readers := [READER_COUNT]Reader{
		Reader{EventId: -1},
		Reader{EventId: -1}}

	rv := &ReaderHandler{ClientHandler: h, Readers: readers}
	rv.Init()
	rv.Name = "readerhandler"
	rv.CoreHandler = core
	return rv
}

func (ah *ReaderHandler) Start() error {
	ah.BaseHandler.Start()

	err := ah.ClientHandler.Connect()
	if err != nil {
		log.Fatal(err)
	}

	defer ah.ClientHandler.Close()

	client := modbus.NewClient(ah.ClientHandler)
	for {
		select {
		case <-ah.Ctx.Done():
			fmt.Println("Context canceled")
			return ah.Ctx.Err()

		default:
			ah.process(client)
		}
	}
}

func (ah *ReaderHandler) process(client modbus.Client) {

	results, _ := client.ReadInputRegisters(
		READER_START_ADDRESS,
		READER_COUNT*READER_BLOCK_SIZE)

	ah.processReaderData(results)

}

func (ah *ReaderHandler) processReaderData(data []byte) {
	if len(data) == 0 {
		return
	}
	for idx, reader := range ah.Readers {
		eventIdPos := 2 * READER_BLOCK_SIZE * idx
		newEventId := int(data[eventIdPos])
		curEventId := reader.EventId
		if curEventId >= 0 && newEventId > 0 {
			if newEventId != curEventId {
				d1 := int(data[eventIdPos+1]) << 16
				d2 := int(data[eventIdPos+2]) << 8
				d3 := int(data[eventIdPos+3])
				ah.Readers[idx].CardCode = d1 | d2 | d3

				t := handler.Tag{
					Name:  fmt.Sprintf("%s%d", "Reader", idx),
					Value: strconv.Itoa(ah.Readers[idx].CardCode)}

				e := handler.Event{
					Source: ah.Name,
					Tag:    t}

				ah.SendEvent(e)
			}
		}
		ah.Readers[idx].EventId = newEventId
	}
}
