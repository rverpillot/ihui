package ihui

import (
	"strings"
)

type Page struct {
	Render
	id      string
	title   string
	actions map[string]ActionFunc
}

func NewPage(id string, title string, render Render) *Page {
	page := &Page{
		id:      id,
		Render:  render,
		title:   title,
		actions: make(map[string]ActionFunc),
	}
	return page
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

func (p *Page) On(name string, action ActionFunc) {
	if !strings.Contains(name, ".") {
		name = p.Id() + "." + name
	}
	p.actions[name] = action
}

func (p *Page) Trigger(name string, ctx *Context) {
	if !strings.Contains(name, ".") {
		name = p.Id() + "." + name
	}
	action := p.actions[name]
	if action == nil {
		return
	}
	action(ctx)
}
