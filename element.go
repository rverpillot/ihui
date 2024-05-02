package ihui

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"

	"github.com/PuerkitoBio/goquery"
)

type Template interface {
	Execute(w io.Writer, model interface{}) error
}

type HTMLRendererFunc func(*HTMLElement) error

func (f HTMLRendererFunc) Render(page *HTMLElement) error { return f(page) }

type HTMLRenderer interface {
	Render(*HTMLElement) error
}

type Options struct {
	Title   string
	Target  string
	Replace bool
	Hide    bool
	Page    bool
}

type HTMLElement struct {
	Id        string
	session   *Session
	renderer  HTMLRenderer
	buffer    bytes.Buffer
	doc       *goquery.Document
	options   Options
	actions   []Action
	templates map[string]Template
	active    bool
}

func newHTMLElement(id string, renderer HTMLRenderer, options Options) *HTMLElement {
	if options.Target == "" {
		options.Target = "body"
	}
	return &HTMLElement{
		Id:        id,
		renderer:  renderer,
		options:   options,
		templates: make(map[string]Template),
	}
}

func (p *HTMLElement) Title() string {
	return p.options.Title
}

func (p *HTMLElement) SetTitle(title string) {
	p.options.Title = title
}

func (p *HTMLElement) Session() *Session {
	return p.session
}

func (p *HTMLElement) IsActive() bool {
	return p.active
}

func (p *HTMLElement) IsPage() bool {
	return p.options.Page
}

func (p *HTMLElement) IsVisible() bool {
	return !p.options.Hide
}

func (p *HTMLElement) ClearCache() {
	p.buffer.Reset()
	p.doc = nil
	p.templates = make(map[string]Template)
}

func (p *HTMLElement) Write(data []byte) (int, error) {
	p.doc = nil
	return p.buffer.Write(data)
}

func (p *HTMLElement) WriteString(html string) {
	p.Write([]byte(html))
}

func (p *HTMLElement) Printf(format string, args ...interface{}) {
	p.WriteString(fmt.Sprintf(format, args...))
}

func (p *HTMLElement) ExecuteTemplate(tpl Template, model any) error {
	return tpl.Execute(p, model)
}

func (p *HTMLElement) WriteGoTemplateString(tpl string, model any) error {
	return p.ExecuteTemplate(NewGoTemplate(p.Id, tpl), model)
}

func (p *HTMLElement) WriteGoTemplate(fsys fs.FS, filename string, model any) error {
	template, ok := p.templates[filename]
	if !ok {
		template = NewGoTemplateFile(fsys, filename)
		p.templates[filename] = template
	}
	return p.ExecuteTemplate(template, model)
}

func (p *HTMLElement) SetHtml(selector string, renderer HTMLRenderer) error {

	doc := p.doc
	if doc == nil {
		var err error
		doc, err = goquery.NewDocumentFromReader(&p.buffer)
		if err != nil {
			return err
		}
	}
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if err := renderer.Render(p); err != nil {
			return
		}
		html, err := p.toHtml()
		if err != nil {
			return
		}
		s.SetHtml(html)
	})
	html, err := doc.Find("body").Html()
	if err != nil {
		return err
	}
	p.buffer.Reset()
	p.WriteString(html)
	p.doc = doc
	return nil
}

func (p *HTMLElement) UniqueId(prefix string) string {
	return p.session.UniqueId(prefix)
}

func (p *HTMLElement) Get(name string) interface{} {
	return p.session.Get(name)
}

// Register an action
func (p *HTMLElement) On(eventName string, selector string, action ActionCallback) {
	if action == nil {
		return
	}
	// log.Printf("Element '%s': Register action %s on %s", p.Id, eventName, selector)
	p.actions = append(p.actions, Action{Name: eventName, Selector: selector, Fct: action})
}

func (p *HTMLElement) OnClick(selector string, action ActionCallback) {
	p.On("click", selector, action)
}

func (p *HTMLElement) OnSubmit(selector string, action ActionCallback) {
	p.On("submit", selector, action)
}

func (p *HTMLElement) OnInput(selector string, action ActionCallback) {
	p.On("input", selector, action)
}

func (p *HTMLElement) OnCheck(selector string, action ActionCallback) {
	p.On("check", selector, action)
}

func (p *HTMLElement) OnForm(selector string, action ActionCallback) {
	p.On("form", selector, action)
}

func (p *HTMLElement) sendEvent(name string, data any) error {

	if p.session == nil {
		return fmt.Errorf("element %s has no session", p.Id)
	}
	return p.session.SendEvent(&Event{
		Name:    name,
		Element: p.Id,
		Target:  p.options.Target,
		Data:    data,
	})
}

// trigger an event. Return true if the event was handled.
func (p *HTMLElement) trigger(event Event) error {
	idAction := -1
	if event.Target == "" {
		for id, action := range p.actions {
			if action.Name == event.Name && action.Selector == "" {
				idAction = id
				break
			}
		}
	} else {
		fmt.Sscanf(event.Target, "action-%d", &idAction)
	}
	if idAction < 0 || idAction >= len(p.actions) {
		return nil
	}
	// log.Printf("Element '%s' - execute: %+v", p.Id, event)
	action := p.actions[idAction]
	return action.Fct(p.session, event)
}

// draw the element
func (p *HTMLElement) draw() error {
	p.actions = nil
	p.buffer.Reset()
	display := "none"
	if !p.options.Hide {
		display = "inline"
	}
	class := "ihui-element"
	if p.options.Page {
		class = "ihui-page"
	}
	p.WriteString(fmt.Sprintf(`<div id="%s" class="%s" style="display: %s">`, p.Id, class, display))
	if p.renderer != nil {
		p.doc = nil
		if err := p.renderer.Render(p); err != nil {
			return err
		}
	}
	p.WriteString("</div>")
	html, err := p.toHtml()
	if err != nil {
		return err
	}

	// log.Printf("Draw element %s", p.Name)
	err = p.sendEvent("element", map[string]interface{}{
		"title":   p.Title(),
		"page":    p.options.Page,
		"html":    html,
		"replace": p.options.Replace,
	})
	if err != nil {
		return err
	}
	p.active = true
	return nil
}

// Close the page and remove it from the session. The page can't be used anymore.
func (p *HTMLElement) Close() error {
	p.active = false
	p.buffer.Reset()
	if p.session != nil {
		p.session.remove(p)
	}
	return p.sendEvent("remove", map[string]interface{}{
		"page": p.options.Page,
	})
}

// Show the page
func (p *HTMLElement) Show() error {
	p.options.Hide = false
	if p.active {
		return p.sendEvent("show", nil)
	}
	return nil
}

// Hide the page
func (p *HTMLElement) Hide() error {
	p.options.Hide = true
	if p.active {
		return p.sendEvent("hide", nil)
	}
	return nil
}

func (page *HTMLElement) toHtml() (string, error) {
	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return "", err
	}

	addAction := func(s *goquery.Selection, name string, evname string, idAction int) {
		value := fmt.Sprintf(`ihui.on(event,"%s","%s","action-%d",this);`, evname, page.Id, idAction)
		s.SetAttr(name, value)
	}

	for id, action := range page.actions {
		if action.Selector == "" {
			continue
		}
		doc.Find(action.Selector).Each(func(i int, s *goquery.Selection) {
			switch action.Name {
			case "click":
				addAction(s, "onclick", action.Name, id)

			case "check":
				addAction(s, "onchange", action.Name, id)

			case "change":
				addAction(s, "onchange", action.Name, id)

			case "input":
				addAction(s, "oninput", action.Name, id)

			case "submit":
				addAction(s, "onsubmit", action.Name, id)
				s.SetAttr("method", "post")
				s.SetAttr("action", "")

			case "form":
				s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
					addAction(ss, "onchange", action.Name, id)
				})
			}
		})
	}

	return doc.Find("body").Html()
}
