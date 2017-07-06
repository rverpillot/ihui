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

func (s *Session) ShowPage(page *Page) (*Event, error) {
	s.page = page
	event, err := page.show(false)
	if err != nil {
		return nil, err
	}
	page.Trigger(event.Source, event.Name, s)
	return event, nil
}

func (s *Session) Script(format string, args ...interface{}) error {
	event := &Event{
		Name:   "script",
		Source: "",
		Data:   fmt.Sprintf(format, args...),
	}

	return s.page.sendEvent(event)
}
