package ihui

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/websocket"
)

type RenderFunc func(*Page)

func (f RenderFunc) Draw(page *Page) { f(page) }

type Render interface {
	Draw(*Page)
}

type Event struct {
	Name   string
	Source string
	Data   interface{}
}

type Params map[string]interface{}

type Session struct {
	Params
	ws   *websocket.Conn
	page *Page
}

func newSession(ws *websocket.Conn) *Session {
	return &Session{
		Params: make(map[string]interface{}),
		ws:     ws,
	}
}

func (s *Session) Page() *Page {
	return s.page
}

func (s *Session) sendEvent(event *Event) error {
	// log.Println(event.Name, event.Source)
	err := websocket.WriteJSON(s.ws, event)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Session) NewPage(id string, title string, render Render) *Page {
	page := &Page{
		session: s,
		id:      id,
		Render:  render,
		title:   title,
	}
	return page
}

func (s *Session) Script(format string, args ...interface{}) error {
	event := &Event{
		Name:   "script",
		Source: "",
		Data:   fmt.Sprintf(format, args...),
	}

	return s.sendEvent(event)
}

type Context struct {
	params map[string]interface{}
	ws     *websocket.Conn
	Event  *Event
}

func newContext(ws *websocket.Conn) *Context {
	return &Context{
		ws:     ws,
		params: make(map[string]interface{}),
	}
}

func (ctx *Context) Set(name string, value interface{}) {
	ctx.params[name] = value
}

func (ctx *Context) Get(name string) interface{} {
	return ctx.params[name]
}

func (ctx *Context) sendEvent(event *Event) error {
	// log.Println(event.Name, event.Source)
	err := websocket.WriteJSON(ctx.ws, event)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (ctx *Context) NewPage(id string, title string, render Render) *Page {
	page := &Page{
		ctx:    ctx,
		id:     id,
		Render: render,
		title:  title,
	}
	return page
}

func (ctx *Context) Script(format string, args ...interface{}) error {
	event := &Event{
		Name:   "script",
		Source: "",
		Data:   fmt.Sprintf(format, args...),
	}

	return ctx.sendEvent(event)
}

type ActionFunc func(*Context)

type HTTPHandler struct {
	Path          string
	CSSPaths      []string
	JSPaths       []string
	Title         string
	box           *rice.Box
	assetHandler  http.Handler
	assetHandler2 http.Handler
	startFunc     ActionFunc
}

func NewHTTPHandler(title string, startFunc ActionFunc) *HTTPHandler {
	box := rice.MustFindBox("ihui_resources")
	return &HTTPHandler{
		Title:        title,
		box:          box,
		assetHandler: http.FileServer(box.HTTPBox()),
		startFunc:    startFunc,
	}
}

func (h *HTTPHandler) AddCss(path string) {
	h.CSSPaths = append(h.CSSPaths, path)
}

func (h *HTTPHandler) AddJs(path string) {
	h.JSPaths = append(h.JSPaths, path)
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Path == "" {
		h.Path = strings.TrimRight(r.URL.Path, "/")
	}
	r.URL.Path = strings.TrimPrefix(r.URL.Path, h.Path)
	log.Println(r.URL.Path)

	if r.URL.Path == "/ws" {
		var upgrader = websocket.Upgrader{}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		context := newContext(ws)
		context.Set("contextRoot", h.Path)
		context.Event = &Event{Name: "init"}
		h.startFunc(context)

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
