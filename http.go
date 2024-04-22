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

func welcomePage(page *Page) error {
	page.WriteString(`
	<section class="section">
		<div class="content">
			<h1>Welcome to ihui</h1>
			<p>ihui is a simple and lightweight web framework for Go.</p>
			<p>It provides a way to build web applications using Go and HTML templates.</p>
			<p>ihui uses websocket to read events and write html response.</p>
		</div>
	</section>
	`)
	return nil
}

type HTTPHandler struct {
	startFunc func(*Session) error
	templ     *template.Template
	Path      string
}

func NewHTTPHandler(startFunc func(*Session) error) http.Handler {
	if startFunc == nil {
		startFunc = func(s *Session) error {
			s.ShowPage("welcome", HTMLRendererFunc(welcomePage), &Options{Title: "Welcome"})
			return nil
		}
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
