package ihui

import (
	"github.com/gorilla/websocket"
)

type Session struct {
	params map[string]interface{}
	ws     *websocket.Conn
	page   *BufferedPage
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

func (s *Session) ShowPage(title string, drawer PageDrawer) error {
	s.page = newPage(title)
	html, err := s.page.render(drawer)
	if err != nil {
		return err
	}

	event := &Event{
		Name:   "update",
		Source: "_main",
		Data: map[string]interface{}{
			"title": s.page.Title(),
			"html":  html,
		},
	}
	err = s.sendEvent(event)
	if err != nil {
		return err
	}
	event, err = s.recvEvent()
	if err != nil {
		return err
	}

	s.page.Trigger(event.Source, event.Name, s)
	return nil
}

func (s *Session) Script(script string) error {
	event := &Event{
		Name: "update",
		Data: script,
	}
	if err := s.sendEvent(event); err != nil {
		return err
	}
	if _, err := s.recvEvent(); err != nil {
		return err
	}

	return nil
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
