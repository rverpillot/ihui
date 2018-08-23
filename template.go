package ihui

import (
	"html/template"
	"log"
)

type PageTemplate struct {
	template *template.Template
	model    interface{}
}

func NewPageTemplate(tmpl string, model interface{}) *PageTemplate {
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		log.Print(err)
		return nil
	}
	return &PageTemplate{
		template: t,
		model:    model,
	}
}

func (p *PageTemplate) SetModel(model interface{}) {
	p.model = model
}

func (p *PageTemplate) Render(page Page) {
	err := p.template.Execute(page, p.model)
	if err != nil {
		log.Print(err)
	}
}
