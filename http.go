package ihui

import (
	"html/template"
	"log"
	"net/http"
	"path"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
)

//go:generate go-bindata-assetfs -pkg ihui -prefix resources/ resources/...

type HTTPHandler struct {
	assetHandler http.Handler
	startFunc    func(*Session)
	templ        *template.Template
	Path         string
}

func NewHTTPHandler(startFunc func(*Session)) *HTTPHandler {
	return &HTTPHandler{
		assetHandler: http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/"}),
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

		content, err := Asset(filename)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write(content)
	}
}
