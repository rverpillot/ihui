package ihui

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

type Options struct {
	Title   string
	Modal   bool
	Target  string
	Visible bool
}

type HTMLRendererFunc func(*Page)

func (f HTMLRendererFunc) Render(page *Page) { f(page) }

type HTMLRenderer interface {
	Render(*Page)
}

type Page struct {
	Id       string
	renderer HTMLRenderer
	buffer   bytes.Buffer
	options  Options
	session  *Session
	actions  []Action
	active   bool
}

func newPage(id string, renderer HTMLRenderer, session *Session, options Options) *Page {
	if options.Target == "" {
		options.Target = "body"
	}
	page := &Page{
		Id:       id,
		renderer: renderer,
		options:  options,
		session:  session,
	}
	return page
}

func (p *Page) Title() string {
	return p.options.Title
}

func (p *Page) SetTitle(title string) {
	p.options.Title = title
}

func (p *Page) Session() *Session {
	return p.session
}

func (p *Page) Modal() bool {
	return p.options.Modal
}

func (p *Page) Write(data []byte) (int, error) {
	return p.buffer.Write(data)
}

func (p *Page) WriteString(html string) {
	p.Write([]byte(html))
}

func (p *Page) Read(data []byte) (int, error) {
	return p.buffer.Read(data)
}

func (p *Page) Reset() {
	p.buffer.Reset()
}

func (p *Page) Add(selector string, part HTMLRenderer) error {
	p.Reset()
	p.renderer.Render(p)
	doc, err := goquery.NewDocumentFromReader(p)
	if err != nil {
		return err
	}
	p.Reset()
	part.Render(p)
	html, err := p.toHtml()
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

func (p *Page) UniqueId(prefix string) string {
	return p.session.UniqueId(prefix)
}

func (p *Page) Get(name string) interface{} {
	return p.session.Get(name)
}

// Register an action
func (p *Page) On(eventName string, selector string, action ActionCallback) {
	if action == nil {
		return
	}
	p.actions = append(p.actions, Action{Name: eventName, Selector: selector, Fct: action})
}

// trigger an event. Return true if the event was handled.
func (p *Page) trigger(event Event) bool {
	idAction := -1
	if event.Target == "page" {
		for id, action := range p.actions {
			if action.Name == event.Name && action.Selector == "page" {
				idAction = id
				break
			}
		}
	} else {
		fmt.Sscanf(event.Target, "action-%d", &idAction)
	}
	if idAction == -1 {
		return false
	}
	// log.Printf("Execute %+v", event)
	p.actions[idAction].Fct(p.session, event)
	return true
}

// Draw the page
func (p *Page) Draw() error {
	p.actions = nil
	p.buffer.Reset()
	display := "none"
	if p.options.Visible {
		display = "inline"
	}
	p.WriteString(fmt.Sprintf(`<div id="%s" class="page" style="display: %s">`, p.Id, display))
	p.renderer.Render(p)
	p.WriteString("</div>")
	html, err := p.toHtml()
	if err != nil {
		return err
	}

	// log.Printf("Draw page %s", p.Name)
	err = p.session.SendEvent(&Event{
		Name:   "page",
		Page:   p.Id,
		Target: p.options.Target,
		Data: map[string]interface{}{
			"title": p.Title(),
			"html":  html,
		},
	})
	if err != nil {
		return err
	}
	p.active = true
	return nil
}

// Send a remove-page event to the client
func (p *Page) remove() error {
	p.active = false
	p.Reset()
	return p.session.SendEvent(&Event{
		Name:   "remove-page",
		Page:   p.Id,
		Target: p.options.Target,
	})
}

// Close the page and remove it from the session. The page can't be used anymore.
func (p *Page) Close() error {
	return p.session.RemovePage(p)
}

// Show the page
func (p *Page) Show() error {
	p.options.Visible = true
	if p.active {
		return p.session.SendEvent(&Event{
			Name:   "show-page",
			Page:   p.Id,
			Target: p.options.Target,
		})
	}
	return nil
}

// Hide the page
func (p *Page) Hide() error {
	p.options.Visible = false
	if p.active {
		return p.session.SendEvent(&Event{
			Name:   "hide-page",
			Page:   p.Id,
			Target: p.options.Target,
		})
	}
	return nil
}

func (page *Page) toHtml() (string, error) {
	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return "", err
	}

	addAction := func(s *goquery.Selection, name string, evname string, pageId string, idAction int) {
		s.SetAttr(name, fmt.Sprintf(`ihui.on(event,"%s","%s","action-%d",this);`, evname, pageId, idAction))
	}

	for id, action := range page.actions {
		if action.Selector == "page" {
			continue
		}
		doc.Find(action.Selector).Each(func(i int, s *goquery.Selection) {
			switch action.Name {
			case "click":
				addAction(s, "onclick", action.Name, page.Id, id)

			case "check":
				addAction(s, "onchange", action.Name, page.Id, id)

			case "change":
				addAction(s, "onchange", action.Name, page.Id, id)

			case "input":
				addAction(s, "oninput", action.Name, page.Id, id)

			case "submit":
				addAction(s, "onsubmit", action.Name, page.Id, id)
				s.SetAttr("method", "post")
				s.SetAttr("action", "")

			case "form":
				s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
					addAction(ss, "onchange", action.Name, page.Id, id)
				})
			}
		})
	}

	return doc.Find("body").Html()
}
