package ihui

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Session struct {
	params map[string]interface{}
	ws     *websocket.Conn
	page   *PageHTML
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

func (s *Session) ShowPage(name string, drawer PageRenderer, options *Options) error {
	if options == nil {
		options = &Options{}
	}
	if s.page == nil {
		options.Modal = true
	}

	page := newHTMLPage(name, drawer, s, *options)

	for !page.exit {
		html, err := page.Render()
		if err != nil {
			log.Print(err)
			return err
		}

		html = fmt.Sprintf(`<div id="%s" class="page">%s</div>`, name, html)

		event := &Event{
			Name: "page",
			Data: map[string]interface{}{
				"name":  name,
				"title": page.Title(),
				"html":  html,
			},
		}
		err = s.sendEvent(event)
		if err != nil {
			log.Print(err)
			return err
		}

		s.page = page
		// if !page.options.Modal {
		// 	return nil
		// }

		err = s.WaitEvent(page)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return s.RemovePage(page)
}

func (s *Session) WaitEvent(page *PageHTML) error {
	for {
		event, err := s.recvEvent()
		if err != nil {
			return err
		}

		if page.Trigger(*event) > 0 {
			break
		}
	}
	return nil
}

func (s *Session) RemovePage(page *PageHTML) error {
	event := &Event{
		Name:   "remove",
		Target: page.Name,
	}
	if err := s.sendEvent(event); err != nil {
		return err
	}
	return nil
}

func (s *Session) Script(script string, args ...interface{}) error {
	event := &Event{
		Name: "script",
		Data: fmt.Sprintf(script, args...),
	}
	if err := s.sendEvent(event); err != nil {
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
