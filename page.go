package ihui

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

type Page interface {
	Title() string
	SetTitle(title string)
	Draw(r PageDrawer)
	WriteString(html string)
	Write(data []byte) (int, error)
	UniqueId(string) string
	On(id string, name string, action ActionFunc)
}

type PageDrawerFunc func(Page)

func (f PageDrawerFunc) Draw(page Page) { f(page) }

type PageDrawer interface {
	Draw(Page)
}

type BufferedPage struct {
	buffer  bytes.Buffer
	title   string
	countID int
	exit    bool
	evt     string
	actions map[string][]Action
}

func newPage(title string) *BufferedPage {
	page := &BufferedPage{
		title:   title,
		countID: 1000,
		evt:     "new",
	}
	return page
}

func (p *BufferedPage) Title() string {
	return p.title
}

func (p *BufferedPage) SetTitle(title string) {
	p.title = title
}

func (p *BufferedPage) Draw(r PageDrawer) {
	r.Draw(p)
}

func (p *BufferedPage) WriteString(html string) {
	p.buffer.WriteString(html)
}

func (p *BufferedPage) Write(data []byte) (int, error) {
	return p.buffer.Write(data)
}

func (p *BufferedPage) UniqueId(prefix string) string {
	p.countID++
	return fmt.Sprintf("%s%d", prefix, p.countID)
}

func (p *BufferedPage) On(name string, selector string, action ActionFunc) {
	if action == nil {
		return
	}
	id := selector
	if name != "load" {
		id = p.UniqueId("a")
	}
	p.actions[id] = append(p.actions[id], Action{Name: name, Selector: selector, Fct: action})
}

func (p *BufferedPage) Trigger(id string, session *Session, value interface{}) int {
	count := 0
	actions, ok := p.actions[id]
	if ok {
		for _, action := range actions {
			action.Fct(session, value)
			count++
		}
	}
	return count
}

func (page *BufferedPage) render(drawer PageDrawer) (string, error) {
	page.actions = make(map[string][]Action)
	page.countID = 0

	page.buffer.Reset()

	if page.evt == "update" {
		page.buffer.WriteString(`<div id="main">`) // because morphdom processing
	}

	if drawer != nil {
		drawer.Draw(page)
	}

	if page.evt == "update" {
		page.buffer.WriteString(`</div>`)
	}

	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return "", err
	}

	addAttr := func(s *goquery.Selection, name string, id string, value string) string {
		attr := "_" + name + "_id"
		id = s.AttrOr(attr, id)
		s.SetAttr(attr, id)
		s.SetAttr(name, fmt.Sprintf(`_s(event,"%s",%s);`, id, value))
		return id
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
				val := s.AttrOr("data-value", s.AttrOr("data-id", s.AttrOr("id", "")))
				_id = addAttr(s, "onclick", id, `"`+val+`"`)
				if goquery.NodeName(s) == "a" {
					s.SetAttr("href", "")
				}

			case "check":
				_id = addAttr(s, "onchange", id, `$(this).prop("checked")`)

			case "change":
				_id = addAttr(s, "onchange", id, `$(this).val()`)

			case "input":
				_id = addAttr(s, "oninput", id, `$(this).val()`)

			case "submit":
				_id = addAttr(s, "onsubmit", id, `$(this).serializeObject()`)

			case "form":
				s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
					_id = addAttr(ss, "onchange", id, `{ name: $(this).attr("name"), val: $(this).val() }`)
				})
			}
			if _id != id {
				page.actions[_id] = append(page.actions[_id], action)
			}
		})
	}

	removeAllAttrs(doc, "_onclick_id", "_onchange_id", "_oninput_id", "_onsubmit_id")

	return doc.Html()
}
