package ihui

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	sessionTimeout = 10 * time.Minute
)

var (
	sessions = make(map[string]*Session)
)

type Session struct {
	id            string
	date          time.Time
	params        map[string]interface{}
	ws            *websocket.Conn
	pages         []*PageHTML
	currentId     int64
	noPageRefresh bool
}

func purgeOldSessions() {
	now := time.Now()
	for _, session := range sessions {
		if session.date.Add(sessionTimeout).Before(now) {
			session.close()
		}
	}
}

func GetSession(id string) *Session {
	purgeOldSessions()
	if id == "" {
		return nil
	}
	if session, ok := sessions[id]; ok {
		return session
	} else {
		return nil
	}
}

func newSession() *Session {
	session := &Session{
		id:        uuid.New().String(),
		date:      time.Now(),
		params:    make(map[string]interface{}),
		currentId: 0,
	}
	sessions[session.id] = session
	return session
}

func (s *Session) close() {
	s.pages = nil
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

func (s *Session) CurrentPage() *PageHTML {
	if len(s.pages) == 0 {
		return nil
	}
	return s.pages[len(s.pages)-1]
}

func (s *Session) UniqueId(prefix string) string {
	s.currentId++
	return fmt.Sprintf("%s%d", prefix, s.currentId)
}

func (s *Session) ShowPage(name string, drawer PageRenderer, options *Options) {
	s.date = time.Now()
	if options == nil {
		options = &Options{}
	}

	page := newHTMLPage(name, drawer, s, *options)

	if page.options.Modal || len(s.pages) == 0 {
		s.pages = append(s.pages, page)
	} else {
		s.pages[len(s.pages)-1] = page
	}
}

func (s *Session) run() error {
	for {
		s.date = time.Now()

		page := s.CurrentPage()
		if page == nil {
			break
		}

		html, err := page.Render()
		if err != nil {
			log.Print(err)
			return err
		}

		html = fmt.Sprintf(`<div id="%s" class="page" style="display: none">%s</div>`, page.Name, html)

		event := &Event{
			Name: "page",
			Data: map[string]interface{}{
				"name":  page.Name,
				"title": page.Title(),
				"html":  html,
			},
		}

		// log.Printf("Display page %s", s.page.Name)
		err = s.SendEvent(event)
		if err != nil {
			return err
		}

		for {
			// log.Print("Wait event")
			event, err = s.RecvEvent()
			if err != nil {
				s.date = time.Now()
				return err
			}

			// log.Printf("Event: %+v\n", event)

			s.noPageRefresh = false
			if page.Trigger(*event) && (event.Refresh && !s.noPageRefresh) {
				break
			}
		}
	}
	return nil
}

func (s *Session) sendRemovePageEvent(page *PageHTML) error {
	event := &Event{
		Name:   "remove",
		Target: page.Name,
	}
	if err := s.SendEvent(event); err != nil {
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
	if err := s.SendEvent(event); err != nil {
		return err
	}
	return nil
}

func (s *Session) ClosePage() {
	page := s.CurrentPage()
	if page != nil {
		s.sendRemovePageEvent(page)
		s.pages = s.pages[:len(s.pages)-1] // remove last page
	}
}

func (s *Session) CloseModalPage() bool {
	page := s.CurrentPage()
	if page != nil && page.options.Modal {
		s.sendRemovePageEvent(page)
		s.pages = s.pages[:len(s.pages)-1] // remove last page
		return true
	}
	return false
}

func (s *Session) PreventPageRefresh() {
	s.noPageRefresh = true
}

func (s *Session) SendEvent(event *Event) error {
	if err := s.ws.WriteJSON(event); err != nil {
		return err
	}
	return nil
}

func (s *Session) RecvEvent() (*Event, error) {
	var event Event
	if err := s.ws.ReadJSON(&event); err != nil {
		return nil, err
	}
	return &event, nil
}
