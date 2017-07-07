package ihui

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/websocket"
)

type Event struct {
	Name   string
	Source string
	Data   interface{}
}

type ActionFunc func(*Session)

// type ActionFunc func(*Context)

type HTTPHandler struct {
	ContextRoot  string
	CSSPaths     []string
	JSPaths      []string
	box          *rice.Box
	assetHandler http.Handler
	startFunc    ActionFunc
}

func NewHTTPHandler(contextroot string, startFunc ActionFunc) *HTTPHandler {
	if strings.HasSuffix(contextroot, "/") {
		contextroot = contextroot[:len(contextroot)-1]
	}

	box := rice.MustFindBox("ihui_resources")
	return &HTTPHandler{
		ContextRoot:  contextroot,
		box:          box,
		assetHandler: http.FileServer(box.HTTPBox()),
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
			t, err := template.New("main").Parse(h.box.MustString("index.html"))
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
