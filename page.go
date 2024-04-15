package ihui

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

type Options struct {
	Title string
	Modal bool
}

type Page interface {
	Write(data []byte) (int, error)
	WriteString(html string)
	Read(data []byte) (int, error)
	Close() error
	Add(selector string, render PageRenderer) error
	SetTitle(title string)
	On(id string, name string, action ActionFunc)
	Session() *Session
	Update(selector string, html string) error
	Get(name string) interface{}
	UniqueId(prefix string) string
}

type PageRendererFunc func(Page)

func (f PageRendererFunc) Render(page Page) { f(page) }

type PageRenderer interface {
	Render(Page)
}

type PageHTML struct {
	Name    string
	drawer  PageRenderer
	buffer  bytes.Buffer
	options Options
	session *Session
	actions []Action
}

func newHTMLPage(name string, drawer PageRenderer, session *Session, options Options) *PageHTML {
	page := &PageHTML{
		Name:    name,
		drawer:  drawer,
		options: options,
		session: session,
	}
	return page
}

func (p *PageHTML) Title() string {
	return p.options.Title
}

func (p *PageHTML) SetTitle(title string) {
	p.options.Title = title
}

func (p *PageHTML) Session() *Session {
	return p.session
}

func (p *PageHTML) Modal() bool {
	return p.options.Modal
}

func (p *PageHTML) Actions() []Action {
	return p.actions
}

func (p *PageHTML) Write(data []byte) (int, error) {
	return p.buffer.Write(data)
}

func (p *PageHTML) WriteString(html string) {
	p.Write([]byte(html))
}

func (p *PageHTML) Read(data []byte) (int, error) {
	return p.buffer.Read(data)
}

func (p *PageHTML) Reset() {
	p.buffer.Reset()
}

func (p *PageHTML) Close() error {
	return nil
}

func (p *PageHTML) Add(selector string, part PageRenderer) error {
	doc, err := goquery.NewDocumentFromReader(p)
	if err != nil {
		return err
	}
	html, err := p.toHtml(part)
	if err != nil {
		return err
	}
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		s.SetHtml(html)
	})
	p.Reset()
	html, _ = doc.Find("body").Html()
	p.WriteString(html)
	return nil
}

func (p *PageHTML) UniqueId(prefix string) string {
	return p.session.UniqueId(prefix)
}

func (p *PageHTML) Get(name string) interface{} {
	return p.session.Get(name)
}

// Register an action
func (p *PageHTML) On(eventName string, selector string, action ActionFunc) {
	if action == nil {
		return
	}
	p.actions = append(p.actions, Action{Name: eventName, Selector: selector, Fct: action})
}

// Trigger an event. Return true if the event was handled.
func (p *PageHTML) Trigger(event Event) bool {
	numAction := -1
	if event.Target == "page" {
		for i, action := range p.actions {
			if action.Name == event.Name && action.Selector == "page" {
				numAction = i
				break
			}
		}
	} else {
		fmt.Sscanf(event.Target, "action-%d", &numAction)
	}
	if numAction < 0 || numAction >= len(p.actions) {
		return false
	}
	// log.Printf("Execute %+v", event)
	p.actions[numAction].Fct(p.session, event)
	return true
}

func (p *PageHTML) Render() (string, error) {
	p.resetActions()
	return p.toHtml(p.drawer)
}

// Update a part of the page
func (p *PageHTML) Update(selector string, html string) error {
	event := &Event{Name: "update", Target: selector, Data: html}
	if err := p.session.SendEvent(event); err != nil {
		return err
	}
	return nil
}

func (p *PageHTML) resetActions() {
	p.actions = nil
}

func (page *PageHTML) toHtml(pageRenderer PageRenderer) (string, error) {
	page.buffer.Reset()

	if pageRenderer != nil {
		pageRenderer.Render(page)
	}

	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return "", err
	}

	addAction := func(s *goquery.Selection, name string, evname string, numAction int) {
		s.SetAttr(name, fmt.Sprintf(`ihui.on(event,"%s","action-%d",this);`, evname, numAction))
	}

	for num, action := range page.actions {
		if action.Selector == "page" {
			continue
		}
		doc.Find(action.Selector).Each(func(i int, s *goquery.Selection) {
			switch action.Name {
			case "click":
				addAction(s, "onclick", action.Name, num)

			case "check":
				addAction(s, "onchange", action.Name, num)

			case "change":
				addAction(s, "onchange", action.Name, num)

			case "input":
				addAction(s, "oninput", action.Name, num)

			case "submit":
				addAction(s, "onsubmit", action.Name, num)
				s.SetAttr("method", "post")
				s.SetAttr("action", "")

			case "form":
				s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
					addAction(ss, "onchange", action.Name, num)
				})
			}
		})
	}

	return doc.Find("body").Html()
}
