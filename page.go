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
	UniqueId() string
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
	actions map[string]Action
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

func (p *BufferedPage) UniqueId() string {
	p.countID++
	return fmt.Sprintf("u%d", p.countID)
}

func (p *BufferedPage) On(name string, selector string, action ActionFunc) {
	if action == nil {
		return
	}
	id := p.UniqueId()
	p.actions[id] = Action{Name: name, Selector: selector, Fct: action}
}

func (p *BufferedPage) Trigger(id string, session *Session) {
	action, ok := p.actions[id]
	if ok {
		action.Fct(session)
	}
}

func (page *BufferedPage) render(drawer PageDrawer) (string, error) {
	page.actions = make(map[string]Action)
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

	for id, action := range page.actions {
		doc.Find(action.Selector).Each(func(i int, s *goquery.Selection) {

			switch action.Name {
			case "click":
				val := s.AttrOr("data-value", s.AttrOr("data-id", s.AttrOr("id", "")))
				s.SetAttr("onclick", `sendMsg(event, "click","`+id+`", "`+val+`")`)
				if goquery.NodeName(s) == "a" {
					s.SetAttr("href", "")
				}

			case "check":
				s.SetAttr("onchange", `sendMsg(event, "check","`+id+`", $(this).prop("checked"))`)

			case "change":
				s.SetAttr("onchange", `sendMsg(event, "change","`+id+`", $(this).val())`)

			case "input":
				s.SetAttr("oninput", `sendMsg(event, "change","`+id+`", $(this).val())`)

			case "submit":
				s.SetAttr("onsubmit", `sendMsg(event, "form","`+id+`", $(this).serializeObject())`)

			case "form":
				s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
					ss.SetAttr("onchange", `sendMsg(event, "change","`+id+`", { name: $(this).attr("name"), val: $(this).val() })`)
				})
			}
		})

	}

	return doc.Html()
}
