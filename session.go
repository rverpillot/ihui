package ihui

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	sessionTimeout = 5 * time.Minute
)

var (
	sessions = make(map[string]*Session)
)

type Session struct {
	id        string
	date      time.Time
	params    map[string]interface{}
	pages     []*Page
	ws        *websocket.Conn
	uniqueId  int64
	noRefresh bool

	lock sync.Mutex
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
		id:       uuid.New().String(),
		date:     time.Now(),
		params:   make(map[string]interface{}),
		uniqueId: 0,
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
	s.lock.Lock()
	defer s.lock.Unlock()
	s.params[name] = value
}

func (s *Session) Get(name string) interface{} {
	return s.params[name]
}

func (s *Session) UniqueId(prefix string) string {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.uniqueId++
	return fmt.Sprintf("%s%d", prefix, s.uniqueId)
}

func (s *Session) getPage(id string) *Page {
	for _, page := range s.pages {
		if page.Id == id {
			return page
		}
	}
	return nil
}

func (s *Session) addPage(page *Page) error {
	// log.Printf("Add page '%s'", page.Id)
	if page.IsModal() {
		page.session = s
		return nil
	}
	page.session = s
	s.lock.Lock()
	defer s.lock.Unlock()

	if idx := slices.IndexFunc(s.pages, func(p *Page) bool { return p.Id == page.Id }); idx >= 0 {
		s.pages[idx] = page
	} else {
		s.pages = append(s.pages, page)
	}
	return nil
}

func (s *Session) removePage(page *Page) error {
	// log.Printf("Remove page '%s'", page.Id)
	if page.IsModal() {
		return page.remove()
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	if idx := slices.IndexFunc(s.pages, func(p *Page) bool { return p.Id == page.Id }); idx >= 0 {
		s.pages = slices.Delete(s.pages, idx, idx+1)
		return page.remove()
	}
	return nil
}

func (s *Session) ShowPage(id string, renderer HTMLRenderer, options *Options) error {
	// log.Printf("Show page '%s'", id)
	if options == nil {
		options = &Options{}
	}
	if options.Modal {
		return s.ShowModal(id, renderer, options)
	} else {
		options.Visible = true
		return s.addPage(newPage(id, renderer, *options))
	}
}

func (s *Session) HidePage(id string) error {
	if page := s.getPage(id); page != nil {
		return page.Hide()
	}
	return fmt.Errorf("Page '%s' not found", id)
}

func (s *Session) ShowModal(id string, renderer HTMLRenderer, options *Options) error {
	// log.Printf("Show page '%s'", id)
	if options == nil {
		options = &Options{}
	}
	options.Visible = true
	options.Modal = true
	page := newPage(id, renderer, *options)
	s.addPage(page)
	for {
		s.date = time.Now()

		if err := page.draw(); err != nil {
			log.Printf("Error: %s", err.Error())
		}

		for {
			// log.Print("Wait event")
			event, err := s.RecvEvent()
			if err != nil {
				s.date = time.Now()
				return err
			}
			// log.Printf("Event: %+v\n", event)

			if event.Page != page.Id {
				continue
			}

			s.noRefresh = false
			if err := page.trigger(*event); err != nil {
				log.Printf("Error: %s", err.Error())
				s.ShowError(err)
			}
			if !page.IsActive() {
				return nil
			}
			if event.Refresh && !s.noRefresh {
				break
			}
		}
	}
}

func (s *Session) run() error {
	for {
		s.date = time.Now()
		for _, page := range s.pages {
			if err := page.draw(); err != nil {
				log.Printf("Error: %s", err.Error())
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

			page := s.getPage(event.Page)
			if page == nil {
				continue
			}

			s.noRefresh = false
			if err := page.trigger(*event); err != nil {
				log.Printf("Error: %s", err.Error())
				s.ShowError(err)
			}
			if event.Refresh && !s.noRefresh {
				break
			}
		}
	}
}

func (s *Session) PreventRefresh() {
	s.noRefresh = true
}

func (s *Session) SendEvent(event *Event) error {
	s.lock.Lock()
	defer s.lock.Unlock()
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
func (s *Session) Execute(script string, args ...interface{}) error {
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

func (s *Session) ShowError(err error) {
	s.Execute(`alert("%s")`, err.Error())
}
