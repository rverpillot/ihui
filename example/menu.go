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
	page.WriteString(`<div id="menu">`)
	for _, name := range menu.names {
		if name == menu.active {
			page.WriteString(fmt.Sprintf(`<div><p>%s</p></div>`, name))
			continue
		}
		id := page.UniqueId("m")
		page.WriteString(fmt.Sprintf(`<div><a id="%s" href="">%s</a></div>`, id, name))
		active := name
		page.On("click", fmt.Sprintf("[id=%s]", id), func(session *ihui.Session, _ ihui.Event) {
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
		menu.pages[name].Render(page)
		page.WriteString(`</div>`)
	}
	return nil
}
