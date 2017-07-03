package ihui

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/websocket"
)

type RenderFunc func(io.Writer)

func (f RenderFunc) Render(w io.Writer) { f.Render(w) }

type Render interface {
	Draw(io.Writer)
}

type Event struct {
	Name   string
	Source string
	Data   interface{}
}

type Context struct {
	Session map[string]interface{}
	params  map[string]interface{}
	ws      *websocket.Conn
	Event   *Event
}

func newContext(ws *websocket.Conn) *Context {
	return &Context{
		ws:      ws,
		params:  make(map[string]interface{}),
		Session: make(map[string]interface{}),
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

func (ctx *Context) Display(r Render, w io.Writer) {
	r.Draw(w)
}

func (ctx *Context) DisplayPage(page *Page, modal bool) *Event {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(`<div><div id="%s" style="height: 100%%">`, page.Id()))
	ctx.Display(page, &buffer)
	buffer.WriteString(`</div></div>`)

	doc, err := goquery.NewDocumentFromReader(&buffer)
	if err != nil {
		log.Println(err)
		return nil
	}

	doc.Find("[data-action]").Each(func(i int, s *goquery.Selection) {
		_, ok := s.Attr("id")
		if !ok {
			return
		}
	})

	html, err := doc.Find("div").First().Html()
	if err != nil {
		log.Println(err)
		return nil
	}

	event := &Event{
		Name:   "update",
		Source: page.Id(),
		Data: map[string]interface{}{
			"title": page.Title(),
			"html":  html,
		},
	}

	if err := ctx.sendEvent(event); err != nil {
		return nil
	}
	err = websocket.ReadJSON(ctx.ws, ctx.Event)
	if err != nil {
		log.Println(err)
		return nil
	}

	name := ctx.Event.Source + "." + ctx.Event.Name
	page.Trigger(name, ctx)

	return ctx.Event
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
