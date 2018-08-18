package ihui

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
)

//go:generate go-bindata-assetfs -pkg ihui -prefix resources/ resources/...

type Event struct {
	Name   string
	Source string
	Data   interface{}
}

type ActionFunc func(*Session)

func (f ActionFunc) String() string { return fmt.Sprintf("#%p", f) }

type HTTPHandler struct {
	contextRoot  string
	index        string
	assetHandler http.Handler
	startFunc    ActionFunc
}

func NewHTTPHandler(contextroot string, startFunc ActionFunc) *HTTPHandler {
	contextroot = strings.TrimSuffix(contextroot, "/")

	return &HTTPHandler{
		contextRoot:  contextroot,
		assetHandler: http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/"}),
		startFunc:    startFunc,
	}
}

func (h *HTTPHandler) Path() string {
	return h.contextRoot
}

func (h *HTTPHandler) SetIndexPage(index string) {
	h.index = index
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	if !strings.HasPrefix(r.URL.Path, h.contextRoot+"/") {
		http.Redirect(w, r, h.contextRoot+"/", http.StatusTemporaryRedirect)
		return
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, h.contextRoot)

	if r.Header.Get("Upgrade") == "websocket" {
		var upgrader = websocket.Upgrader{}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		session := newSession(ws)
		session.Set("path", h.contextRoot)
		h.startFunc(session)

	} else {
		if r.URL.Path == "/" {
			index := h.index
			if index == "" {
				index = string(MustAsset("index.html"))
			}
			t, err := template.New("main").Parse(index)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			t.Execute(w, h)
		} else {
			h.assetHandler.ServeHTTP(w, r)
		}
	}
}
