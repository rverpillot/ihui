package main

import (
	"fmt"

	ihui "rverpi/ihui.v2"
)

type Menu struct {
	names  []string
	pages  map[string]ihui.PageDrawer
	active string
}

func NewMenu() *Menu {
	return &Menu{pages: make(map[string]ihui.PageDrawer)}
}

func (menu *Menu) Add(name string, r ihui.PageDrawer) {
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

func (menu *Menu) Draw(page ihui.Page) {
	page.WriteString(`<div id="menu">`)
	for _, name := range menu.names {
		if name == menu.active {
			page.WriteString(fmt.Sprintf(`<div><p>%s</p></div>`, name))
			continue
		}
		id := page.UniqueId("m")
		page.WriteString(fmt.Sprintf(`<div><a id="%s">%s</a></div>`, id, name))
		active := name
		page.On("click", fmt.Sprintf("[id=%s]", id), func(session *ihui.Session) {
			menu.SetActive(active)
		})
	}
	page.WriteString(`</div>`)
	for _, name := range menu.names {
		style := "display:none"
		if name == menu.active {
			style = ""
		}
		page.WriteString(fmt.Sprintf(`<div id="%s" style="%s">`, name, style))
		page.Draw(menu.pages[name])
		page.WriteString(`</div>`)
	}
}
