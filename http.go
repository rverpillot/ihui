package ihui

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/websocket"
)

//go:embed resources/*
var resources embed.FS

type HTTPHandler struct {
	assetHandler http.Handler
	startFunc    func(*Session) error
	templ        *template.Template
	Path         string
}

func NewHTTPHandler(startFunc func(*Session) error) *HTTPHandler {
	fsys, err := fs.Sub(resources, "resources")
	if err != nil {
		panic(err)
	}
	if startFunc == nil {
		panic("startFunc is nil")
	}
	return &HTTPHandler{
		assetHandler: http.FileServer(http.FS((fsys))),
		startFunc:    startFunc,
		templ:        template.New("main"),
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
		filename := "index.html"
		ext := path.Ext(r.URL.Path)
		if ext == ".js" {
			filename = path.Join("js", path.Base(r.URL.Path))
		}
		if ext == ".css" {
			filename = path.Join("css", path.Base(r.URL.Path))
		}

		// log.Printf("GET %s\n", path.Join("resources", filename))
		content, err := resources.ReadFile(path.Join("resources", filename))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write(content)
	}
}
