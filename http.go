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
	startFunc    func(*Session)
	templ        *template.Template
	Path         string
}

func NewHTTPHandler(startFunc func(*Session)) *HTTPHandler {
	fsys, err := fs.Sub(resources, "resources")
	if err != nil {
		panic(err)
	}
	return &HTTPHandler{
		assetHandler: http.FileServer(http.FS((fsys))),
		startFunc:    startFunc,
		templ:        template.New("main"),
	}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	if r.Header.Get("Upgrade") == "websocket" {
		var upgrader = websocket.Upgrader{
			EnableCompression: true,
		}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		session := newSession(ws)
		session.sendEvent(&Event{Name: "init"})
		h.startFunc(session)
		session.Close()

	} else {
		filename := "index.html"
		ext := path.Ext(r.URL.Path)
		if ext == ".js" {
			filename = path.Join("js", path.Base(r.URL.Path))
		}
		if ext == ".css" {
			filename = path.Join("css", path.Base(r.URL.Path))
		}

		// log.Println(filename)
		content, err := resources.ReadFile(path.Join("resources", filename))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write(content)
	}
}
