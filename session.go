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
	html, err := s.page.Html()
	if err != nil {
		return nil, err
	}

	event := &Event{
		Name:   "update",
		Source: "ihui_main",
		Data: map[string]interface{}{
			"title": s.page.Title(),
			"html":  html,
		},
	}
	err = s.sendEvent(event)
	if err != nil {
		return nil, err
	}
	event, err = s.recvEvent()
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

	return s.sendEvent(event)
}

func (s *Session) sendEvent(event *Event) error {
	if err := websocket.WriteJSON(s.ws, event); err != nil {
		return err
	}
	return nil
}

func (s *Session) recvEvent() (*Event, error) {
	var event Event
	if err := websocket.ReadJSON(s.ws, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
