package main

import (
	"fmt"

	ihui "rverpi90/ihui.v2"
)

type Menu struct {
	names  []string
	pages  map[string]ihui.PageRenderer
	active string
}

func NewMenu() *Menu {
	return &Menu{pages: make(map[string]ihui.PageRenderer)}
}

func (menu *Menu) Add(name string, r ihui.PageRenderer) {
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

func (menu *Menu) Render(page ihui.Page) {
	page.WriteString(`<div id="menu">`)
	for _, name := range menu.names {
		if name == menu.active {
			page.WriteString(fmt.Sprintf(`<div><p>%s</p></div>`, name))
			continue
		}
		id := page.UniqueId("m")
		page.WriteString(fmt.Sprintf(`<div><a id="%s">%s</a></div>`, id, name))
		active := name
		page.On("click", fmt.Sprintf("[id=%s]", id), func(session *ihui.Session, _ ihui.Event) bool {
			menu.SetActive(active)
			return true
		})
	}
	page.WriteString(`</div>`)
	for _, name := range menu.names {
		style := "display:none"
		if name == menu.active {
			style = ""
		}
		page.WriteString(fmt.Sprintf(`<div id="%s" style="%s">`, name, style))
		menu.pages[name].Render(page)
		page.WriteString(`</div>`)
	}
}
