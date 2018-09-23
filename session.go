package ihui

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	sessions = make(map[string]*Session)
)

type Session struct {
	id             string
	params         map[string]interface{}
	refreshPage    bool
	ws             *websocket.Conn
	page           *PageHTML
	countId        map[string]int64
	actionsHistory map[string][]Action
}

func GetSession(id string) *Session {
	return sessions[id]
}

func newSession(ws *websocket.Conn) *Session {
	session := &Session{
		id:      uuid.New().String(),
		params:  make(map[string]interface{}),
		ws:      ws,
		countId: make(map[string]int64),
	}
	sessions[session.id] = session
	return session
}

func (s *Session) Close() {
	s.ws.Close()
	delete(sessions, s.id)
}

func (s *Session) Id() string {
	return s.id
}

func (s *Session) Set(name string, value interface{}) {
	s.params[name] = value
}

func (s *Session) Get(name string) interface{} {
	return s.params[name]
}

func (s *Session) CurrentPage() Page {
	return s.page
}

func (s *Session) UniqueId(prefix string) string {
	count, ok := s.countId[prefix]
	if !ok {
		count = 0
	}
	count++
	s.countId[prefix] = count
	return fmt.Sprintf("%s%d", prefix, count)
}

func (s *Session) ShowPage(name string, drawer PageRenderer, options *Options) bool {
	if options == nil {
		options = &Options{}
	}
	if s.page == nil {
		options.Modal = true
	}

	previous := s.page

	s.page = newHTMLPage(name, drawer, s, *options)
	if s.page.options.Modal {
		if err := s.waitEvent(); err != nil {
			log.Print(err)
			return false
		}
		s.page = previous
	}
	return true
}

func (s *Session) waitEvent() error {
	actionsHistory := make(map[string][]Action)

	for !s.page.exit {
		html, err := s.page.Render()
		if err != nil {
			log.Print(err)
			return err
		}

		html = fmt.Sprintf(`<div id="%s" class="page" style="display: none">%s</div>`, s.page.Name, html)

		event := &Event{
			Name: "page",
			Data: map[string]interface{}{
				"name":  s.page.Name,
				"title": s.page.Title(),
				"html":  html,
			},
		}

		// log.Printf("Display page %s", s.page.Name)
		err = s.sendEvent(event)
		if err != nil {
			return err
		}

		for {
			// log.Print("Wait event")
			event, err = s.recvEvent()
			if err != nil {
				return err
			}

			if n := s.page.Trigger(*event, actionsHistory); n > 0 {
				break
			}
		}
	}

	// log.Printf("Remove page %s", s.page.Name)
	s.RemovePage(s.page)

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

func (s *Session) CloseModalPage() bool {
	if s.page.options.Modal {
		s.page.exit = true
	}
	return s.page.exit
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
