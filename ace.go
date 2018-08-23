package ihui

import (
	"html/template"
	"log"

	"github.com/yosssi/ace"
)

type PageAce struct {
	template *template.Template
	model    interface{}
}

func NewPageAce(content []byte, model interface{}) *PageAce {
	options := ace.InitializeOptions(nil)
	options.Asset = func(name string) ([]byte, error) {
		return content, nil
	}

	t, err := ace.Load("content", "", options)
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

func (p *PageAce) Render(page Page) {
	err := p.template.Execute(page, p.model)
	if err != nil {
		log.Print(err)
	}
}
