package ihui

import (
	"html/template"

	"github.com/yosssi/ace"
)

type PageAce struct {
	template *template.Template
	model    interface{}
}

func NewPageAce(filename string, content []byte, model interface{}) *PageAce {
	options := new(ace.Options)
	if len(content) > 0 {
		options.Asset = func(name string) ([]byte, error) {
			return content, nil
		}
	}
	t, err := ace.Load(filename, "", options)
	if err != nil {
		panic(err)
	}
	return &PageAce{
		template: t,
		model:    model,
	}
}

func (p *PageAce) SetModel(model interface{}) {
	p.model = model
}

func (p *PageAce) Render(page *Page) {
	err := p.template.Execute(page, p.model)
	if err != nil {
		panic(err)
	}
}
