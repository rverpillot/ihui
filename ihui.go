package ihui

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"text/template"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
)

//go:generate go-bindata-assetfs -pkg ihui -prefix resources/ resources/...

type Event struct {
	Name   string
	Id     string
	Target string
	Data   interface{}
}

func (e *Event) Value() string {
	return e.Data.(string)
}

func (e *Event) IsChecked() bool {
	t, ok := e.Data.(bool)
	if !ok {
		return false
	}
	return t
}

type ActionFunc func(*Session, Event) bool

func (f ActionFunc) String() string { return fmt.Sprintf("#%p", f) }

type Action struct {
	Selector string
	Name     string
	Fct      ActionFunc
}

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

		t, err := template.New(filename).Parse(string(content))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		h.Path = path.Dir(r.URL.Path)

		err = t.Execute(w, h)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}
