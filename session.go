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
	page := newPage(title)

	for !page.exit {
		s.page = page

		html, err := page.render(drawer)
		if err != nil {
			return err
		}

		event := &Event{
			Name: page.evt,
			Data: map[string]interface{}{
				"title": page.Title(),
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

		page.Trigger(event.Source, s)

		if s.page != page {
			page.evt = "new"
		} else {
			page.evt = "update"
		}
	}
	return nil
}

func (s *Session) Script(script string) error {
	event := &Event{
		Name: "script",
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

func (s *Session) QuitPage() {
	s.page.exit = true
}
