package handler

type Handler interface {
	Start() error
	Stop()
	GetTags(key string) string
	SetTag(key string, value string)
}

type BaseHandler struct {
	Handler
}

func (bh *BaseHandler) Stop() {

}

type Tag struct {
	Name  string
	Value string
}
