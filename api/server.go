package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tommyblue/her/her"
)

type Intent struct {
	Action string `json:"action"`
	Room   string `json:"room"`
}

type Server struct {
	outCh       chan<- her.Message
	router      *mux.Router
	intentConfs []her.IntentConf
	host        string
	port        int
}

func NewServer(host string, port int, outCh chan her.Message) (*Server, error) {
	if port == 0 {
		port = 8080
	}

	s := &Server{
		outCh: outCh,
		host:  host,
		port:  port,
	}

	if err := viper.UnmarshalKey("intents", &s.intentConfs); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Start() {
	s.router = mux.NewRouter() //.StrictSlash(true)
	s.router.HandleFunc("/", s.homeLink)
	s.router.HandleFunc("/alexa/", s.alexaLink) //.Methods("POST")
	go func() {
		address := fmt.Sprintf("%s:%d", s.host, s.port)
		log.Info("Listening on ", address)
		if err := http.ListenAndServe(address, s.router); err != nil {
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

	s.applyIntent(i)

	fmt.Fprintf(w, "ok")
}

func (s *Server) homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func (s *Server) applyIntent(i Intent) {
	for _, intentConf := range s.intentConfs {
		if intentConf.Action == i.Action && intentConf.Room == i.Room {
			log.Info(fmt.Sprintf("Applying Action: %s, Room: %s", i.Action, i.Room))
			s.outCh <- her.Message{Topic: intentConf.Topic, Message: []byte(intentConf.Message)}
			return
		}
	}
	log.Warning(fmt.Sprintf("Cannot find Action: %s, Room: %s", i.Action, i.Room))
}
