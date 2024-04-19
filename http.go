package ihui

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//go:embed resources/js/ihui.min.js
var js []byte

type HTTPHandler struct {
	startFunc func(*Session) error
	templ     *template.Template
	Path      string
}

func NewHTTPHandler(startFunc func(*Session) error) *HTTPHandler {
	if startFunc == nil {
		panic("startFunc is nil")
	}
	return &HTTPHandler{
		startFunc: startFunc,
		templ:     template.New("main"),
	}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path)

	if r.Header.Get("Upgrade") == "websocket" {
		var upgrader = websocket.Upgrader{
			EnableCompression: true,
		}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer ws.Close()

		var event Event
		if err := ws.ReadJSON(&event); err != nil || event.Name != "connect" {
			log.Println(err)
			return
		}

		var session *Session
		if oldSession := GetSession(event.Id); oldSession != nil && len(oldSession.pages) > 0 {
			log.Printf("Reconnect to session %s\n", oldSession.id)
			session = oldSession
			session.ws = ws
		} else {
			session = newSession()
			session.ws = ws
			session.SendEvent(&Event{Name: "init", Id: session.Id()})
			if err := h.startFunc(session); err != nil {
				log.Println(err)
				session.close()
				return
			}
		}
		if err := session.run(); err != nil {
			log.Println(err)
		}

	} else {
		w.Write(js)
	}
}
