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

func (s *Session) ShowPage(drawer PageRenderer, options *Options) error {
	if options == nil {
		options = &Options{}
	}
	if s.page == nil {
		options.Modal = true
	}

	page := newHTMLPage(s, *options)

	if !page.Modal() {
		s.page = page
		return nil
	}

	for evt := "new"; !page.exit; {
		s.page = page

		html, err := page.Render(drawer)
		if err != nil {
			log.Print(err)
			return err
		}

		if evt == "update" {
			html = fmt.Sprintf(`<div id="main">%s</div>`, html) // morphdom
		}

		event := &Event{
			Name: evt,
			Data: map[string]interface{}{
				"title": page.Title(),
				"html":  html,
			},
		}
		err = s.sendEvent(event)
		if err != nil {
			log.Print(err)
			return err
		}

		for {
			event, err = s.recvEvent()
			if err != nil {
				log.Print(err)
				return err
			}

			if page.Trigger(*event) > 0 {
				break
			}
		}

		if page != s.page {
			if !s.page.Modal() {
				page = s.page
			}
			evt = "new"
		} else {
			evt = "update"
		}
	}
	page.Trigger(Event{Name: "unload", Source: "page", Target: "page"})
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
