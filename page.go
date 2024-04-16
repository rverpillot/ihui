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
}

func newPage(id string, renderer HTMLRenderer, session *Session, options Options) *Page {
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

func (p *Page) Actions() []Action {
	return p.actions
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

func (p *Page) Close() error {
	p.Reset()
	p.session.removePage(p)
	return p.session.SendEvent(&Event{
		Name:   "remove",
		Target: "#pages",
		Data:   map[string]interface{}{"id": p.Id},
	})
}

func (p *Page) Add(selector string, part HTMLRenderer) error {
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

func (p *Page) UniqueId(prefix string) string {
	return p.session.UniqueId(prefix)
}

func (p *Page) Get(name string) interface{} {
	return p.session.Get(name)
}

// Register an action
func (p *Page) On(eventName string, selector string, action ActionFunc) {
	if action == nil {
		return
	}
	p.actions = append(p.actions, Action{Name: eventName, Selector: selector, Fct: action})
}

// Trigger an event. Return true if the event was handled.
func (p *Page) Trigger(event Event) bool {
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

// Draw the page
func (p *Page) Draw() error {
	p.buffer.Reset()
	p.actions = nil
	p.WriteString(fmt.Sprintf(`<div id="%s" class="page" style="display: none">`, p.Id))
	html, err := p.toHtml(nil)
	if err != nil {
		return err
	}
	p.WriteString("</div>")

	// log.Printf("Draw page %s", p.Name)
	return p.session.SendEvent(&Event{
		Name:   "page",
		Target: "#pages",
		Data: map[string]interface{}{
			"id":    p.Id,
			"title": p.Title(),
			"html":  html,
		},
	})
}

// Show the page
func (p *Page) Show() {
	p.session.showPage(p)
}

func (page *Page) toHtml(pageRenderer HTMLRenderer) (string, error) {
	if pageRenderer != nil {
		pageRenderer.Render(page)
	} else {
		page.renderer.Render(page)
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
