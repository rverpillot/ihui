package ihui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PageDrawerFunc func(*Page)

func (f PageDrawerFunc) Draw(page *Page) { f(page) }

type PageDrawer interface {
	Draw(*Page)
}

type Page struct {
	PageDrawer
	buffer  bytes.Buffer
	title   string
	countID int
	actions map[string][]ActionFunc
}

func NewPage(title string, render PageDrawer) *Page {
	page := &Page{
		PageDrawer: render,
		title:      title,
	}
	return page
}

func NewPageFunc(title string, fct PageDrawerFunc) *Page {
	return NewPage(title, PageDrawerFunc(fct))
}

func (p *Page) Title() string {
	return p.title
}

func (p *Page) SetTitle(title string) {
	p.title = title
}

func (p *Page) Render(r PageDrawer) {
	r.Draw(p)
}

func (p *Page) WriteString(html string) {
	p.buffer.WriteString(html)
}

func (p *Page) Write(data []byte) {
	p.buffer.Write(data)
}

func (p *Page) NewId() string {
	p.countID++
	return fmt.Sprintf("i%d", p.countID)
}

func (p *Page) On(id string, name string, action ActionFunc) {
	if action == nil {
		return
	}
	name = id + "." + name
	p.actions[name] = append(p.actions[name], action)
}

func (p *Page) Trigger(id string, name string, session *Session) {
	name = id + "." + name
	actions := p.actions[name]
	for _, action := range actions {
		action(session)
	}
}

func (page *Page) Html() (string, error) {
	page.actions = make(map[string][]ActionFunc)
	page.countID = 0

	page.buffer.Reset()
	page.buffer.WriteString(`<div id="ihui_main">`)
	page.Draw(page)
	page.buffer.WriteString(`</div>`)

	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return "", err
	}

	doc.Find("[id]").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("id")

		for name := range page.actions {
			if strings.HasPrefix(name, id+".") {
				action := strings.Split(name, ".")[1]

				switch action {
				case "click":
					s.SetAttr("onclick", `sendMsg(event, "click", $(this).attr("id"), null)`)
					if goquery.NodeName(s) == "a" {
						s.SetAttr("href", "")
					}

				case "check":
					s.SetAttr("onchange", `sendMsg(event, "check", $(this).attr("id"), $(this).prop("checked"))`)

				case "change":
					s.SetAttr("onchange", `sendMsg(event, "change", $(this).attr("id"), $(this).val())`)

				case "input":
					s.SetAttr("oninput", `sendMsg(event, "change", $(this).attr("id"), $(this).val())`)

				case "submit":
					s.SetAttr("onsubmit", `sendMsg(event, "form", $(this).attr("id"), $(this).serializeObject())`)

				case "form":
					s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
						ss.SetAttr("onchange", `sendMsg(event, "change", $(this).attr("id"), { name: $(this).attr("name"), val: $(this).val() })`)
					})
				}
			}

		}

	})

	return doc.Html()
}
