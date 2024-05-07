package main

import (
	"github.com/rverpillot/ihui"
)

type Item struct {
	Name     string
	r        ihui.HTMLRenderer
	IsActive bool
}

type Menu struct {
	Items []*Item
}

func NewMenu() *Menu {
	return &Menu{}
}

func (menu *Menu) AddItem(name string, r ihui.HTMLRenderer) {
	menu.Items = append(menu.Items, &Item{Name: name, r: r, IsActive: false})
	if len(menu.Items) == 1 {
		menu.Items[0].IsActive = true
	}
}

func (menu *Menu) SetActiveItem(name string) {
	for _, item := range menu.Items {
		item.IsActive = (name == item.Name)
	}
}

func (menu *Menu) ActiveItem() *Item {
	for _, item := range menu.Items {
		if item.IsActive {
			return item
		}
	}
	return nil
}

func (menu *Menu) Render(e *ihui.HTMLElement) error {
	tmpl := `
	<section class="section">
		<div class="tabs">
			<ul>
			{{range .Items}}
				<li {{if .IsActive}}class="is-active"{{end}}><a id="{{.Name}}" href="">{{.Name}}</a></li>
			{{end}}
			</ul>
		</div>
	</section>
	<section class="section">
		<div id="content"></div>
	</section>
	`

	e.Session().AddElement("content", menu.ActiveItem().r)

	e.OnClick("a", func(s *ihui.Session, e ihui.Event) error {
		menu.SetActiveItem(e.Id)
		return nil
	})

	return e.WriteGoTemplateString(tmpl, menu)
}
