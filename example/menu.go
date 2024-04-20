package main

import (
	"github.com/rverpillot/ihui"
	"github.com/rverpillot/ihui/templating"
)

type Item struct {
	Name     string
	r        ihui.HTMLRenderer
	IsActive bool
}

type Menu struct {
	tmpl  *templating.PageMustache
	Items []*Item
}

func NewMenu() *Menu {
	return &Menu{
		tmpl: templating.NewPageMustache(`
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
		`),
	}
}

func (menu *Menu) Add(name string, r ihui.HTMLRenderer) {
	menu.Items = append(menu.Items, &Item{Name: name, r: r, IsActive: false})
	if len(menu.Items) == 1 {
		menu.SetActive(name)
	}
}

func (menu *Menu) SetActive(name string) {
	for _, item := range menu.Items {
		item.IsActive = item.Name == name
	}
}

func (menu *Menu) Active() ihui.HTMLRenderer {
	for _, item := range menu.Items {
		if item.IsActive {
			return item.r
		}
	}
	return nil
}

func (menu *Menu) Render(page *ihui.Page) error {
	if err := page.WriteTemplate(menu.tmpl, menu); err != nil {
		return err
	}

	if err := page.Include("[data-id=content]", menu.Active()); err != nil {
		return err
	}

	page.On("click", "a", func(s *ihui.Session, e ihui.Event) error {
		menu.SetActive(e.Id)
		return nil
	})
	return nil
}
