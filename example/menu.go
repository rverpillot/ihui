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
		menu.SetActiveItem(name)
	}
}

func (menu *Menu) SetActiveItem(name string) {
	for _, item := range menu.Items {
		item.IsActive = item.Name == name
	}
}

func (menu *Menu) ActiveItem() ihui.HTMLRenderer {
	for _, item := range menu.Items {
		if item.IsActive {
			return item.r
		}
	}
	return nil
}

func (menu *Menu) Render(page *ihui.Page) error {
	tmpl := `
	<section class="section">
		<div class="tabs">
			<ul>
			{{#Items}}
				<li {{#IsActive}}class="is-active"{{/IsActive}}><a id="{{Name}}" href="">{{Name}}</a></li>
			{{/Items}}
			</ul>
		</div>
	</section>
	<section class="section" data-id="content">
	</section>
	`
	if err := page.WriteMustacheString(tmpl, menu); err != nil {
		return err
	}

	if err := page.SetHtml("[data-id=content]", menu.ActiveItem()); err != nil {
		return err
	}

	page.On("click", "a", func(s *ihui.Session, e ihui.Event) error {
		menu.SetActiveItem(e.Id)
		return nil
	})
	return nil
}
