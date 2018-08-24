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
	WriteString(html string)
	Write(data []byte) (int, error)
	Add(string, PageRenderer) error
	UniqueId(string) string
	SetTitle(string)
	Session() *Session
	Get(string) interface{}
	On(id string, name string, action ActionFunc)
	Script(string, ...interface{}) error
}

type PageRendererFunc func(Page)

func (f PageRendererFunc) Render(page Page) { f(page) }

type PageRenderer interface {
	Render(Page)
}

type PageHTML struct {
	buffer  bytes.Buffer
	options Options
	title   string
	countID int
	exit    bool
	session *Session
	actions map[string][]Action
}

func newHTMLPage(session *Session, options Options) *PageHTML {
	page := &PageHTML{
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

func (p *PageHTML) Add(selector string, render PageRenderer) error {
	doc, err := goquery.NewDocumentFromReader(&p.buffer)
	if err != nil {
		return err
	}
	html, err := p.html(render)
	if err != nil {
		return err
	}
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		s.SetHtml(html)
	})
	p.buffer.Reset()
	html, _ = doc.Find("body").Html()
	p.buffer.WriteString(html)
	return nil
}

func (p *PageHTML) WriteString(html string) {
	p.buffer.WriteString(html)
}

func (p *PageHTML) Write(data []byte) (int, error) {
	return p.buffer.WriteString(string(data))
}

func (p *PageHTML) UniqueId(prefix string) string {
	p.countID++
	return fmt.Sprintf("%s%d", prefix, p.countID)
}

func (p *PageHTML) Get(name string) interface{} {
	return p.session.Get(name)
}

func (p *PageHTML) On(name string, selector string, action ActionFunc) {
	if action == nil {
		return
	}
	id := selector
	if id != "page" {
		id = p.UniqueId("a")
	}
	p.actions[id] = append(p.actions[id], Action{Name: name, Selector: selector, Fct: action})
}

func (p *PageHTML) Trigger(event Event) int {
	count := 0
	actions, ok := p.actions[event.Target]
	if ok {
		for _, action := range actions {
			if action.Name != event.Name {
				continue
			}
			action.Fct(p.session, event)
			count++
		}
	}
	if event.Name == "load" {
		count = 0
	}
	return count
}

func (p *PageHTML) Script(script string, args ...interface{}) error {
	return p.session.Script(script, args...)
}

func (p *PageHTML) Render(drawer PageRenderer) (string, error) {
	p.actions = make(map[string][]Action)
	p.countID = 1000
	return p.html(drawer)
}

func (page *PageHTML) html(drawer PageRenderer) (string, error) {
	page.buffer.Reset()

	if drawer != nil {
		drawer.Render(page)
	}

	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return "", err
	}

	addAction := func(s *goquery.Selection, name string, target string, value string) string {
		source := s.AttrOr("data-id", s.AttrOr("id", ""))
		attr := "_action_id"
		target = s.AttrOr(attr, target)
		s.SetAttr(attr, target)
		s.SetAttr("on"+name, fmt.Sprintf(`trigger(event,"%s","%s","%s",%s);`, name, source, target, value))
		return target
	}

	removeAllAttrs := func(doc *goquery.Document, names ...string) {
		for _, name := range names {
			doc.Find("[" + name + "]").Each(func(i int, s *goquery.Selection) {
				s.RemoveAttr(name)
			})
		}
	}

	for id, actions := range page.actions {
		action := actions[0]

		if action.Name == "load" {
			continue
		}

		doc.Find(action.Selector).Each(func(i int, s *goquery.Selection) {
			_id := id

			switch action.Name {
			case "click":
				value := s.AttrOr("data-value", s.AttrOr("data-id", s.AttrOr("id", "")))
				_id = addAction(s, "click", id, `"`+value+`"`)
				if goquery.NodeName(s) == "a" {
					s.SetAttr("href", "")
				}

			case "check":
				_id = addAction(s, "change", id, `$(this).prop("checked")=="checked"`)

			case "change":
				_id = addAction(s, "change", id, `$(this).val()`)

			case "input":
				_id = addAction(s, "input", id, `$(this).val()`)

			case "submit":
				_id = addAction(s, "submit", id, `$(this).serializeObject()`)

			case "form":
				s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
					_id = addAction(ss, "change", id, `{ name: $(this).attr("name"), val: $(this).val() }`)
				})
			}
			if _id != id {
				page.actions[_id] = append(page.actions[_id], action)
			}
		})
	}

	removeAllAttrs(doc, "_action_id")

	return doc.Find("body").Html()
}
