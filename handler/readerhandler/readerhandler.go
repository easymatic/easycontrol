package readerhandler

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

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

func NewReaderHandler() *ReaderHandler {
	handler := modbus.NewRTUClientHandler("/dev/ttyMDB")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second

	readers := [READER_COUNT]Reader{
		Reader{EventId: -1},
		Reader{EventId: -1}}

	return &ReaderHandler{ClientHandler: handler, Readers: readers}
}

func (ah *ReaderHandler) Start(eventchan chan string) {
	fmt.Println("starting reader handler")

	ctx := context.Background()
	ah.BaseHandler.Ctx, ah.BaseHandler.Cancel = context.WithCancel(ctx)

	err := ah.ClientHandler.Connect()
	if err != nil {
		log.Fatal(err)
	}

	defer ah.ClientHandler.Close()

	client := modbus.NewClient(ah.ClientHandler)
	for {
		select {
		case <-ah.BaseHandler.Ctx.Done():
			fmt.Println("Context canceled")
			return

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

				fmt.Printf(
					"ah.reader[%v]: EventId=%v, CardCode=%v \n",
					idx,
					ah.Readers[idx].EventId,
					ah.Readers[idx].CardCode)

				// TODO fire ReadCardCode event
			}
		}
		ah.Readers[idx].EventId = newEventId
	}
}

func (ah *ReaderHandler) GetTags(key string) string {
	return ""
}

func (ah *ReaderHandler) SetTag(key string, value string) {

}
