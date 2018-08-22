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
	Title() string
	SetTitle(title string)
	Draw(r PageDrawer)
	WriteString(html string)
	Write(data []byte) (int, error)
	UniqueId(string) string
	Get(string) interface{}
	On(id string, name string, action ActionFunc)
}

type PageDrawerFunc func(Page)

func (f PageDrawerFunc) Draw(page Page) { f(page) }

type PageDrawer interface {
	Draw(Page)
}

type BufferedPage struct {
	buffer  bytes.Buffer
	options Options
	title   string
	countID int
	exit    bool
	evt     string
	session *Session
	drawer  PageDrawer
	actions map[string][]Action
}

func newPage(session *Session, drawer PageDrawer, options Options) *BufferedPage {
	page := &BufferedPage{
		options: options,
		countID: 1000,
		evt:     "new",
		session: session,
		drawer:  drawer,
	}
	return page
}

func (p *BufferedPage) Title() string {
	return p.options.Title
}

func (p *BufferedPage) SetTitle(title string) {
	p.options.Title = title
}

func (p *BufferedPage) Modal() bool {
	return p.options.Modal
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

func (p *BufferedPage) Get(name string) interface{} {
	return p.session.Get(name)
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

func (p *BufferedPage) Trigger(event Event) int {
	count := 0
	actions, ok := p.actions[event.Target]
	if ok {
		for _, action := range actions {
			action.Fct(p.session, event)
			count++
		}
	}
	return count
}

func (page *BufferedPage) render() (string, error) {
	page.actions = make(map[string][]Action)
	page.countID = 0

	page.buffer.Reset()

	if page.evt == "update" {
		page.buffer.WriteString(`<div id="main">`) // because morphdom processing
	}

	if page.drawer != nil {
		page.drawer.Draw(page)
	}

	if page.evt == "update" {
		page.buffer.WriteString(`</div>`)
	}

	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return "", err
	}

	addAction := func(s *goquery.Selection, name string, target string, value string) string {
		source := s.AttrOr("data-id", s.AttrOr("id", ""))
		attr := "_" + name + "_id"
		target = s.AttrOr(attr, target)
		s.SetAttr(attr, target)
		s.SetAttr(name, fmt.Sprintf(`_s(event,"%s","%s","%s",%s);`, name, source, target, value))
		return target
	}

	removeAllAttrs := func(doc *goquery.Document, names ...string) {
		for _, name := range names {
			doc.Find("[" + name + "]").Each(func(i int, s *goquery.Selection) {
				s.RemoveAttr(name)
			})
		}
	}

	/*
		doc.Find("[onclick]").Each(func(i int, s *goquery.Selection) {
			attr, _ := s.Attr("onclick")
			if attr[0] == '#' {
				val := s.AttrOr("data-value", s.AttrOr("data-id", s.AttrOr("id", "")))
				s.SetAttr("onclick", fmt.Sprintf(`_s(even,"%s",%s)`, attr, `"`+val+`"`))
				if goquery.NodeName(s) == "a" {
					s.SetAttr("href", "")
				}
			}
		})
	*/

	for id, actions := range page.actions {
		action := actions[0]

		if action.Name == "load" {
			continue
		}

		doc.Find(action.Selector).Each(func(i int, s *goquery.Selection) {
			_id := id

			switch action.Name {
			case "click":
				_id = addAction(s, "onclick", id, `""`)
				if goquery.NodeName(s) == "a" {
					s.SetAttr("href", "")
				}

			case "check":
				_id = addAction(s, "onchange", id, `$(this).prop("checked")`)

			case "change":
				_id = addAction(s, "onchange", id, `$(this).val()`)

			case "input":
				_id = addAction(s, "oninput", id, `$(this).val()`)

			case "submit":
				_id = addAction(s, "onsubmit", id, `$(this).serializeObject()`)

			case "form":
				s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
					_id = addAction(ss, "onchange", id, `{ name: $(this).attr("name"), val: $(this).val() }`)
				})
			}
			if _id != id {
				page.actions[_id] = append(page.actions[_id], action)
			}
		})
	}

	removeAllAttrs(doc, "_onclick_id", "_onchange_id", "_oninput_id", "_onsubmit_id", "oncheck_id")

	return doc.Html()
}
