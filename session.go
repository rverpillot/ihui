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
	pages          []*PageHTML
	currentId      int64
	preventRefresh bool
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

func (s *Session) Exit() {
	s.Close()
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
		err = s.sendEvent(event)
		if err != nil {
			return err
		}

		for {
			// log.Print("Wait event")
			event, err = s.recvEvent()
			if err != nil {
				s.date = time.Now()
				return err
			}

			// log.Printf("Event: %+v\n", event)

			s.preventRefresh = false
			if page.Trigger(*event) && (event.Refresh && !s.preventRefresh) {
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

func (s *Session) PreventRefresh() {
	s.preventRefresh = true
}

func (s *Session) sendEvent(event *Event) error {
	if err := s.ws.WriteJSON(event); err != nil {
		return err
	}
	return nil
}

func (s *Session) recvEvent() (*Event, error) {
	var event Event
	if err := s.ws.ReadJSON(&event); err != nil {
		return nil, err
	}
	return &event, nil
}
