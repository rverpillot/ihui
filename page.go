package ihui

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

type HTMLRendererFunc func(*Page)

func (f HTMLRendererFunc) Render(page *Page) { f(page) }

type HTMLRenderer interface {
	Render(*Page)
}

type Options struct {
	Title   string
	Modal   bool
	Target  string
	Replace bool
	Visible bool
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

func newPage(id string, renderer HTMLRenderer, options Options) *Page {
	if options.Target == "" {
		options.Target = "body"
	}
	return &Page{
		Id:       id,
		renderer: renderer,
		options:  options,
	}
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

func (p *Page) IsModal() bool {
	return p.options.Modal
}

func (p *Page) Write(data []byte) (int, error) {
	return p.buffer.Write(data)
}

func (p *Page) WriteString(html string) {
	p.Write([]byte(html))
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
	// log.Printf("Page '%s': Register action %s on %s", p.Id, eventName, selector)
	p.actions = append(p.actions, Action{Name: eventName, Selector: selector, Fct: action})
}

func (p *Page) sendEvent(name string, data any) error {
	if p.session == nil {
		return fmt.Errorf("Page %s has no session", p.Id)
	}
	return p.session.SendEvent(&Event{
		Name:   name,
		Page:   p.Id,
		Target: p.options.Target,
		Data:   data,
	})
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
	// log.Printf("Page '%s' - execute: %+v", p.Id, event)
	p.actions[idAction].Fct(p.session, event)
	return true
}

// draw the page
func (p *Page) draw() error {
	p.actions = nil
	p.buffer.Reset()
	display := "none"
	if p.options.Visible {
		display = "inline"
	}
	p.WriteString(fmt.Sprintf(`<div id="%s" class="page" style="display: %s">`, p.Id, display))
	if p.renderer != nil {
		p.renderer.Render(p)
	}
	p.WriteString("</div>")
	html, err := p.toHtml()
	if err != nil {
		return err
	}

	// log.Printf("Draw page %s", p.Name)
	err = p.sendEvent("page", map[string]interface{}{
		"title":   p.Title(),
		"html":    html,
		"replace": p.options.Replace,
	})
	if err != nil {
		return err
	}
	p.active = true
	return nil
}

// Send a remove-page event to the client
func (p *Page) remove() error {
	return p.sendEvent("remove-page", nil)
}

// Close the page and remove it from the session. The page can't be used anymore.
func (p *Page) Close() error {
	p.active = false
	p.buffer.Reset()
	if p.session == nil {
		return nil
	}
	return p.session.removePage(p)
}

// Show the page
func (p *Page) Show() error {
	p.options.Visible = true
	if p.active {
		return p.sendEvent("show-page", nil)
	}
	return nil
}

// Hide the page
func (p *Page) Hide() error {
	p.options.Visible = false
	if p.active {
		return p.sendEvent("hide-page", nil)
	}
	return nil
}

func (p *Page) Add(id string, component HTMLRenderer) *Page {
	// log.Printf("Add component '%s'", id)
	page := newPage(id, component, Options{Visible: true, Target: "#" + p.Id, Replace: true})
	p.session.addPage(page)
	return page
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
