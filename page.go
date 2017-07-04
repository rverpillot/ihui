package ihui

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/websocket"
)

type Page struct {
	Render
	ctx     *Context
	buffer  bytes.Buffer
	id      string
	title   string
	countId int
	actions map[string][]ActionFunc
}

func (p *Page) Id() string {
	return p.id
}

func (p *Page) Title() string {
	return p.title
}

func (p *Page) SetTitle(title string) {
	p.title = title
}

func (p *Page) WriteString(html string) {
	p.buffer.WriteString(html)
}

func (p *Page) Write(data []byte) {
	p.buffer.Write(data)
}

func (p *Page) NewId() string {
	p.countId++
	return fmt.Sprintf("%s.%d", p.Id(), p.countId)
}

func (p *Page) On(id string, name string, action ActionFunc) {
	name = id + "." + name
	p.actions[name] = append(p.actions[name], action)
}

func (p *Page) Trigger(id string, name string, ctx *Context) {
	name = id + "." + name
	actions := p.actions[name]
	for _, action := range actions {
		action(ctx)
	}
}

func (page *Page) Show(modal bool) (*Event, error) {
	page.actions = make(map[string][]ActionFunc)
	page.countId = 0

	page.buffer.Reset()
	page.buffer.WriteString(fmt.Sprintf(`<div><div id="%s" style="height: 100%%">`, page.Id()))
	page.Draw(page)
	page.buffer.WriteString(`</div></div>`)

	doc, err := goquery.NewDocumentFromReader(&page.buffer)
	if err != nil {
		return nil, err
	}

	doc.Find("[data-action]").Each(func(i int, s *goquery.Selection) {
		_, ok := s.Attr("id")
		if !ok {
			return
		}
		// s.SetAttr("id", page.Id()+"."+id)
		action, _ := s.Attr("data-action")
		switch action {
		case "click":
			s.SetAttr("onclick", `sendMsg("click", $(this).attr("id"), null)`)

		case "check":
			s.SetAttr("onchange", `sendMsg("check", $(this).attr("id"), $(this).prop("checked"))`)

		case "change":
			s.SetAttr("onchange", `sendMsg("change", $(this).attr("id"), $(this).val())`)

		case "input":
			s.SetAttr("oninput", `sendMsg("change", $(this).attr("id"), $(this).val())`)

		case "submit":
			s.SetAttr("onsubmit", `sendMsg("form", $(this).attr("id"), $(this).serializeObject())`)

		case "form":
			s.Find("input[name], textarea[name], select[name]").Each(func(i int, ss *goquery.Selection) {
				ss.SetAttr("onchange", `sendMsg("change", $(this).attr("id"), { name: $(this).attr("name"), val: $(this).val() })`)
			})
		}

		s.RemoveAttr("data-action")
	})

	html, err := doc.Html()
	if err != nil {
		return nil, err
	}

	event := &Event{
		Name:   "update",
		Source: page.Id(),
		Data: map[string]interface{}{
			"title": page.Title(),
			"html":  html,
		},
	}

	if err := page.ctx.sendEvent(event); err != nil {
		return nil, err
	}
	err = websocket.ReadJSON(page.ctx.ws, page.ctx.Event)
	if err != nil {
		return nil, err
	}

	page.Trigger(page.ctx.Event.Source, page.ctx.Event.Name, page.ctx)

	return page.ctx.Event, nil
}
