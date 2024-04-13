package ihui

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	sessions = make(map[string]*Session)
)

type Session struct {
	id             string
	date           time.Time
	params         map[string]interface{}
	ws             *websocket.Conn
	page           *PageHTML
	currentId      int64
	actionsHistory map[string][]Action
}

func getSession(id string) *Session {
	if id == "" {
		return nil
	}
	if session, ok := sessions[id]; ok {
		return session
	} else {
		return nil
	}
}

func purgeOldSessions(delay time.Duration) {
	now := time.Now()
	for id, session := range sessions {
		if session.date.Add(delay).Before(now) {
			session.Close()
			delete(sessions, id)
		}
	}
}

func newSession(ws *websocket.Conn) *Session {
	session := &Session{
		id:        uuid.New().String(),
		date:      time.Now(),
		params:    make(map[string]interface{}),
		ws:        ws,
		currentId: 0,
	}
	sessions[session.id] = session
	return session
}

func (s *Session) Close() {
	s.ws.Close()
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
	s.currentId++
	return fmt.Sprintf("%s%d", prefix, s.currentId)
}

func (s *Session) ShowPage(name string, drawer PageRenderer, options *Options) bool {
	s.date = time.Now()
	if options == nil {
		options = &Options{}
	}
	if s.page == nil { // Main page
		options.Modal = true
	}

	previous := s.page

	s.page = newHTMLPage(name, drawer, s, *options)

	if s.page.options.Modal {
		if err := s.display(); err != nil {
			log.Print(err)
			return false
		}
		s.page = previous
	}
	return true
}

func (s *Session) display() error {
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

	s.sendRemovePageEvent(s.page)
	return nil
}

func (s *Session) sendRemovePageEvent(page *PageHTML) error {
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
	s.date = time.Now()
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
	s.date = time.Now()
	if err := s.ws.WriteJSON(event); err != nil {
		return err
	}
	return nil
}

func (s *Session) recvEvent() (*Event, error) {
	s.date = time.Now()
	var event Event
	if err := s.ws.ReadJSON(&event); err != nil {
		return nil, err
	}
	return &event, nil
}
