package ihui

import (
	"html/template"
	"log"
)

type PageTemplate struct {
	template *template.Template
	data     interface{}
}

func NewPageTemplate(tmpl string, data interface{}) *PageTemplate {
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		log.Print(err)
		return nil
	}
	return &PageTemplate{
		template: t,
		data:     data,
	}
}

func (ptr *PageTemplate) Render(page Page) {
	err := ptr.template.Execute(page, ptr.data)
	if err != nil {
		log.Print(err)
	}
}
