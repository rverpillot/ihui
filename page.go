package ihui

import (
	"bytes"
	"fmt"
	"strings"

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
	actions map[string][]ActionFunc
}

func newPage(title string) *BufferedPage {
	page := &BufferedPage{
		title:   title,
		countID: 1000,
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

func (p *BufferedPage) Quit() {
	p.exit = true
}

func (p *BufferedPage) MustQuit() bool {
	return p.exit
}

func (p *BufferedPage) UniqueId() string {
	p.countID++
	return fmt.Sprintf("u%d", p.countID)
}

func (p *BufferedPage) On(id string, name string, action ActionFunc) {
	if action == nil {
		return
	}
	name = id + "/" + name
	p.actions[name] = append(p.actions[name], action)
}

func (p *BufferedPage) Trigger(id string, name string, session *Session) {
	name = id + "/" + name
	actions := p.actions[name]
	for _, action := range actions {
		action(session)
	}
}

func (page *BufferedPage) render(drawer PageDrawer) (string, error) {
	page.actions = make(map[string][]ActionFunc)
	page.countID = 0

	page.buffer.Reset()
	page.buffer.WriteString(`<div id="_main">`)
	if drawer != nil {
		drawer.Draw(page)
	}
	page.buffer.WriteString(`</div>`)

	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return "", err
	}

	doc.Find("[id]").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("id")

		for name := range page.actions {
			if strings.HasPrefix(name, id+"/") {
				action := strings.Split(name, "/")[1]

				switch action {
				case "click":
					s.SetAttr("onclick", `sendMsg(event, "click","`+id+`", null)`)
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
			}

		}

	})

	return doc.Html()
}
