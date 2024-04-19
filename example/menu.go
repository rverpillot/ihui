package main

import (
	"fmt"

	"github.com/rverpillot/ihui"
)

type Menu struct {
	names  []string
	pages  map[string]ihui.HTMLRenderer
	active string
}

func NewMenu() *Menu {
	return &Menu{pages: make(map[string]ihui.HTMLRenderer)}
}

func (menu *Menu) Add(name string, r ihui.HTMLRenderer) {
	menu.names = append(menu.names, name)
	menu.pages[name] = r
	if menu.active == "" {
		menu.active = name
	}
}

func (menu *Menu) SetActive(name string) {
	if _, ok := menu.pages[name]; ok {
		menu.active = name
	}
}

func (menu *Menu) Render(page *ihui.Page) error {
	page.WriteString(`<section class="section">`)
	page.WriteString(`<div class="tabs">`)
	page.WriteString(`<ul>`)
	for _, name := range menu.names {
		var is_active string
		if name == menu.active {
			is_active = "is-active"
		} else {
			is_active = ""
		}
		page.WriteString(fmt.Sprintf(`<li class="%s"><a id="%s" href="">%s</a></li>`, is_active, name, name))
		page.On("click", "#"+name, func(session *ihui.Session, _ ihui.Event) error {
			menu.SetActive(name)
			return nil
		})
	}
	page.WriteString(`</ul>`)
	page.WriteString(`</div>`)
	page.WriteString(`</section>`)
	page.WriteString(`<section class="section">`)
	for _, name := range menu.names {
		style := "display:none"
		if name == menu.active {
			style = ""
		}
		page.WriteString(fmt.Sprintf(`<div id="%s" style="%s">`, name, style))
		menu.pages[name].Render(page)
		page.WriteString(`</div>`)
	}
	page.WriteString(`</section>`)
	return nil
}
