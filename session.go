package ihui

import (
	"fmt"
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
	pages         map[string]*Page
	page_modal    *Page
	ws            *websocket.Conn
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
		pages:     make(map[string]*Page),
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

func (s *Session) UniqueId(prefix string) string {
	s.currentId++
	return fmt.Sprintf("%s%d", prefix, s.currentId)
}

func (s *Session) addPage(page *Page) {
	if page.options.Modal {
		s.page_modal = page
	}
	s.pages[page.Id] = page
}

func (s *Session) removePage(page *Page) {
	if page == s.page_modal {
		s.page_modal = nil
	}
	delete(s.pages, page.Id)
}

func (s *Session) CreatePage(name string, drawer HTMLRenderer, options *Options) *Page {
	if options == nil {
		options = &Options{}
	}
	return newPage(name, drawer, s, *options)
}

func (s *Session) ShowPage(name string, renderer HTMLRenderer, options *Options) {
	s.CreatePage(name, renderer, options).Show()
}

func (s *Session) run() error {
	for {
		s.date = time.Now()

		if s.page_modal != nil {
			if err := s.page_modal.Draw(); err != nil {
				return err
			}
		} else {
			for _, page := range s.pages {
				if err := page.Draw(); err != nil {
					return err
				}
			}
		}

		for {
			// log.Print("Wait event")
			event, err := s.RecvEvent()
			if err != nil {
				s.date = time.Now()
				return err
			}

			// log.Printf("Event: %+v\n", event)

			page, ok := s.pages[event.Page]
			if !ok {
				continue
			}

			// Ignore event if it is not for the modal page
			if s.page_modal != nil && page.Id != s.page_modal.Id {
				continue
			}

			s.noPageRefresh = false
			if page.trigger(*event) && (event.Refresh && !s.noPageRefresh) {
				break
			}
		}
	}
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

// Execute a script on the client side
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

// UpdateHtml directly update the content of the element with the given selector
func (s *Session) UpdateHtml(selector string, html string) error {
	event := &Event{Name: "update", Target: selector, Data: html}
	if err := s.SendEvent(event); err != nil {
		return err
	}
	return nil
}
