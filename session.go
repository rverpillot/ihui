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
	id         string
	date       time.Time
	params     map[string]interface{}
	pages      []*Page
	page_modal *Page
	ws         *websocket.Conn
	uniqueId   int64
	noRefresh  bool

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
	page.session = s
	if page.options.Modal {
		if s.page_modal != nil {
			if err := s.page_modal.remove(); err != nil {
				return err
			}
		}
		s.page_modal = page
		return nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if idx := slices.IndexFunc(s.pages, func(p *Page) bool { return p.Id == page.Id }); idx >= 0 {
		s.pages[idx] = page
	} else {
		s.pages = append(s.pages, page)
	}
	// fmt.Println("--------------------")
	// for i, p := range s.pages {
	// 	fmt.Printf("%d: Id:%s addr:%p actions: %d\n", i, p.Id, p, len(p.actions))
	// }
	return nil
}

func (s *Session) removePage(page *Page) error {
	// log.Printf("Remove page '%s'", page.Id)
	if page == s.page_modal {
		err := s.page_modal.remove()
		s.page_modal = nil
		return err
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
	options.Visible = true
	page := newPage(id, renderer, *options)
	return s.addPage(page)
}

func (s *Session) HidePage(id string) error {
	if page := s.getPage(id); page != nil {
		return page.Hide()
	}
	return fmt.Errorf("Page '%s' not found", id)
}

func (s *Session) run() error {
	for {
		s.date = time.Now()

		if s.page_modal != nil {
			if err := s.page_modal.draw(); err != nil {
				log.Printf("Error: %s", err.Error())
			}
		} else {
			for _, page := range s.pages {
				if err := page.draw(); err != nil {
					log.Printf("Error: %s", err.Error())
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

			var page *Page
			if s.page_modal == nil {
				page = s.getPage(event.Page)
			} else {
				if event.Page == s.page_modal.Id { // Ignore event if it is not for the modal page
					page = s.page_modal
				}
			}

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

func (s *Session) ShowError(err error) {
	s.Script(`alert("%s")`, err.Error())
}
