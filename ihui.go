package ihui

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
)

//go:generate go-bindata-assetfs -pkg ihui -prefix ihui_resources/ ihui_resources/...

type Event struct {
	Name   string
	Source string
	Data   interface{}
}

type ActionFunc func(*Session)

type HTTPHandler struct {
	ContextRoot  string
	CSSPaths     []string
	JSPaths      []string
	assetHandler http.Handler
	startFunc    ActionFunc
}

func NewHTTPHandler(contextroot string, startFunc ActionFunc) *HTTPHandler {
	contextroot = strings.TrimSuffix(contextroot, "/")

	return &HTTPHandler{
		ContextRoot:  contextroot,
		assetHandler: http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/"}),
		startFunc:    startFunc,
	}
}

func (h *HTTPHandler) Pattern() string {
	return h.ContextRoot + "/"
}

func (h *HTTPHandler) AddCss(path string) {
	h.CSSPaths = append(h.CSSPaths, path)
}

func (h *HTTPHandler) AddJs(path string) {
	h.JSPaths = append(h.JSPaths, path)
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.TrimPrefix(r.URL.Path, h.ContextRoot)
	log.Println(r.URL.Path)

	if r.URL.Path == "/ws" {
		var upgrader = websocket.Upgrader{}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		session := newSession(ws)
		session.Set("contextroot", h.ContextRoot)
		h.startFunc(session)

	} else {
		if r.URL.Path == "/" {
			t, err := template.New("main").Parse(string(MustAsset("index.html")))
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
