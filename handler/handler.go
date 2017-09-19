package handler

type Handler interface {
	Start() error
	Stop()
	GetTags(key string) string
	SetTag(key string, value string)
}
