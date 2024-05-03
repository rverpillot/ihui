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
	sessionTimeout = 1 * time.Minute
)

var (
	sessions = make(map[string]*Session)
)

type Session struct {
	id        string
	date      time.Time
	params    map[string]interface{}
	page      *HTMLElement
	elements  []*HTMLElement
	ws        *websocket.Conn
	uniqueId  int64
	noRefresh bool

	lock sync.Mutex
}

func purgeOldSessions() {
	now := time.Now()
	for _, session := range sessions {
		if session.date.Add(sessionTimeout).Before(now) {
			session.Close()
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

func (s *Session) Close() {
	if s.IsClosed() {
		return
	}
	s.page = nil
	s.elements = nil
	s.ws = nil
	delete(sessions, s.id)
}

func (s *Session) IsClosed() bool {
	return s.ws == nil
}

func (s *Session) Quit() error {
	s.SendEvent(&Event{Name: "quit"})
	s.Close()
	return nil
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

func (s *Session) getElement(id string) *HTMLElement {
	if s.page != nil && s.page.Id == id {
		return s.page
	}
	for _, e := range s.elements {
		if e.Id == id {
			return e
		}
	}
	return nil
}

func (s *Session) Refresh(ws *websocket.Conn) {
	s.ws = ws
	if s.page != nil {
		s.page.ClearCache()
	}
	for _, page := range s.elements {
		page.ClearCache()
	}
}

func (s *Session) add(element *HTMLElement) error {
	// log.Printf("Add element '%s'", page.Id)
	element.session = s
	if element.IsPage() {
		s.page = element
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	if idx := slices.IndexFunc(s.elements, func(e *HTMLElement) bool { return e.Id == element.Id }); idx >= 0 {
		s.elements[idx] = element
	} else {
		s.elements = append(s.elements, element)
	}
	return nil
}

func (s *Session) remove(element *HTMLElement) error {
	// log.Printf("Remove element '%s'", element.Id)
	if element.IsPage() {
		if s.page != nil && s.page.Id == element.Id {
			s.page = nil
		}
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	if idx := slices.IndexFunc(s.elements, func(e *HTMLElement) bool { return e.Id == element.Id }); idx >= 0 {
		s.elements = slices.Delete(s.elements, idx, idx+1)
	}
	return nil
}

func (s *Session) ShowPage(id string, renderer HTMLRenderer, options *Options) error {
	// log.Printf("Show page '%s'", id)
	if options == nil {
		options = &Options{}
	}
	options.Page = true
	return s.add(newHTMLElement(id, renderer, *options))
}

func (s *Session) AddElement(id string, renderer HTMLRenderer, options *Options) error {
	if options == nil {
		options = &Options{}
	}
	options.Page = false
	return s.add(newHTMLElement(id, renderer, *options))
}

func (s *Session) Show(id string) error {
	if element := s.getElement(id); element != nil {
		return element.Show()
	}
	return fmt.Errorf("element '%s' not found", id)
}

func (s *Session) Hide(id string) error {
	if element := s.getElement(id); element != nil {
		return element.Hide()
	}
	return fmt.Errorf("element '%s' not found", id)
}

func (s *Session) ShowModal(id string, renderer HTMLRenderer, options *Options) error {
	// log.Printf("Show page '%s'", id)
	if options == nil {
		options = &Options{}
	}
	previous_page := s.page
	defer func() {
		s.page = previous_page
	}()
	if err := s.ShowPage(id, renderer, options); err != nil {
		return err
	}
	return s.run(true)
}

func (s *Session) run(modal bool) error {
	defer func() { s.date = time.Now() }()

	for !s.IsClosed() {
		if s.page != nil {
			if err := s.page.draw(); err != nil {
				log.Printf("Error: %s", err.Error())
			}
		}

		for _, e := range s.elements {
			if err := e.draw(); err != nil {
				log.Printf("Error: %s", err.Error())
			}
		}

		for {
			// log.Print("Wait event")
			event, err := s.RecvEvent()
			if err != nil {
				return err
			}
			// log.Printf("Event: %+v\n", event)

			element := s.getElement(event.Element)
			if element == nil {
				continue
			}
			if modal && s.page.Id != event.Element {
				continue
			}

			s.noRefresh = false
			if err := element.trigger(*event); err != nil {
				log.Printf("Error: %s", err.Error())
				s.ShowError(err)
			}
			if s.IsClosed() {
				return nil
			}
			if event.Refresh && !s.noRefresh {
				break
			}
		}
	}
	return nil
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
