package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/tommyblue/her/her"
)

type Intent struct {
	Action string `json:"action"`
	Room   string `json:"room"`
}

type Server struct {
	outCh  chan<- her.Message
	router *mux.Router
}

func Start(outCh chan her.Message) {
	s := Server{
		outCh: outCh,
	}
	s.Run()
}

func (s *Server) Run() {
	s.router = mux.NewRouter() //.StrictSlash(true)
	s.router.HandleFunc("/", s.homeLink)
	s.router.HandleFunc("/alexa/", s.alexaLink) //.Methods("POST")
	go func() {
		if err := http.ListenAndServe(":8080", s.router); err != nil {
			log.Error(err)
		}
	}()
}

func (s *Server) alexaLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome alexa!")

	var i Intent
	err := json.NewDecoder(r.Body).Decode(&i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("Action: %s, Room: %s", i.Action, i.Room))
	if i.Action == "switch-on" && i.Room == "salotto" {
		s.outCh <- her.Message{Topic: "cmnd/sf-salotto/Power", Message: []byte("ON")}
	}
	fmt.Fprintf(w, "ok")
}

func (s *Server) homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}
