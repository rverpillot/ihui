package ihui

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Session struct {
	params map[string]interface{}
	ws     *websocket.Conn
	page   *Page
}

func newSession(ws *websocket.Conn) *Session {
	return &Session{
		params: make(map[string]interface{}),
		ws:     ws,
	}
}

func (s *Session) Set(name string, value interface{}) {
	s.params[name] = value
}

func (s *Session) Get(name string) interface{} {
	return s.params[name]
}

func (s *Session) Page() *Page {
	return s.page
}

func (s *Session) NewPage(title string, render Renderer) *Page {
	page := &Page{
		ws:       s.ws,
		Renderer: render,
		title:    title,
	}
	return page
}

func (s *Session) NewPageFunc(title string, fct RenderFunc) *Page {
	return s.NewPage(title, RenderFunc(fct))
}

func (s *Session) ShowPage(page *Page) (*Event, error) {
	s.page = page
	event, err := page.show(false)
	if err != nil {
		return nil, err
	}
	page.Trigger(event.Source, event.Name, s)
	return event, nil
}

func (s *Session) ShowNewPage(title string, render Renderer) (*Event, error) {
	return s.ShowPage(s.NewPage(title, render))
}

func (s *Session) ShowNewPageFunc(title string, fct RenderFunc) (*Event, error) {
	return s.ShowPage(s.NewPageFunc(title, fct))
}

func (s *Session) Script(format string, args ...interface{}) error {
	event := &Event{
		Name:   "script",
		Source: "",
		Data:   fmt.Sprintf(format, args...),
	}

	return s.page.sendEvent(event)
}
