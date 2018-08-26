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

	if s.page != nil {
		log.Printf("Save previous page %s", s.page.Name)
	}
	previous := s.page
	s.page = newHTMLPage(name, drawer, s, *options)
	if s.page.options.Modal {
		return s.WaitEvent(previous)
	}
	return nil
}

func (s *Session) WaitEvent(previous *PageHTML) error {

	for !s.page.exit {
		html, err := s.page.Render()
		if err != nil {
			log.Print(err)
			return err
		}

		html = fmt.Sprintf(`<div id="%s" class="page">%s</div>`, s.page.Name, html)

		event := &Event{
			Name: "page",
			Data: map[string]interface{}{
				"name":  s.page.Name,
				"title": s.page.Title(),
				"html":  html,
			},
		}

		log.Printf("Display page %s", s.page.Name)
		err = s.sendEvent(event)
		if err != nil {
			log.Print(err)
			return err
		}

		for {
			log.Print("Wait event")
			event, err = s.recvEvent()
			if err != nil {
				return err
			}

			if s.page.Trigger(*event) > 0 {
				break
			}
		}
	}

	log.Printf("Remove page %s", s.page.Name)
	s.RemovePage(s.page)

	if previous != nil {
		log.Printf("Back to page %s", previous.Name)
		s.page = previous
	}
	return nil
}

func (s *Session) RemovePage(page *PageHTML) error {
	event := &Event{
		Name:   "remove",
		Target: page.Name,
	}
	if err := s.sendEvent(event); err != nil {
		log.Print(err)
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
	if s.page.options.Modal {
		s.page.exit = true
	}
}
