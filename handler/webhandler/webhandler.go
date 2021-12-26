package webhandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/easymatic/easycontrol/handler"
)

type WebHandler struct {
	handler.BaseHandler
}

func NewWebHandler(core handler.CoreHandler) *WebHandler {
	web := WebHandler{}
	web.Name = "webhandler"
	web.CoreHandler = core
	return &web
}
func (hndl *WebHandler) hndlr(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("unable to read body")
	}
	log.Infof("polling: %s", string(body))
	cmd := handler.Command{
		Destination: "plchandler",
		Tag: handler.Tag{
			Name:  "ylic",
			Value: "1"}}
	hndl.CoreHandler.RunCommand(cmd)
}

type resp struct {
	Active bool `json:"active"`
}

func (hndl *WebHandler) sensor(w http.ResponseWriter, r *http.Request) {

	log.Infof("method: %s", r.Method)
	req := resp{}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("unable to read body")
	}
	log.Infof("sensor: %s", string(body))
	if err := json.Unmarshal(body, &req); err != nil {
		log.WithError(err).Error("can't decode")
		return
	}
	log.Infof("have event: %+v", req)
	response := struct {
		Active bool `json:"active"`
	}{
		Active: true,
	}
	json.NewEncoder(w).Encode(response)
}

type errorResponse struct {
	Message string
}

type response struct {
	Source string
	Tag    string
	Value  string
}

func (hndl *WebHandler) handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if r.Method == "GET" {
		t, err := hndl.CoreHandler.GetTag(vars["handler"], vars["tag"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := response{
			Source: vars["handler"],
			Tag:    t.Name,
			Value:  t.Value,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}
	if r.Method == "POST" {
		req := response{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("unable to parse json: %s", err.Error()), http.StatusBadRequest)
			return
		}
		cmd := handler.Command{
			Destination: req.Source,
			Tag: handler.Tag{
				Name:  req.Tag,
				Value: req.Value,
			},
		}
		hndl.CoreHandler.RunCommand(cmd)
		resp := response{
			Source: vars["handler"],
			Tag:    req.Tag,
			Value:  req.Value,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func (hndl *WebHandler) Start() error {
	r := mux.NewRouter()
	r.HandleFunc("/handlers/{handler}/{tag}", hndl.handle)
	http.Handle("/", r)
	hndl.BaseHandler.Start()
	// http.HandleFunc("/sensor", hndl.sensor)
	// http.HandleFunc("/", hndl.hndlr)
	log.Fatal(http.ListenAndServe(":8000", nil))
	return nil
}
