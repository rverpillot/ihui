package ihui

import (
	"html/template"
	"log"
)

type TemplateDrawer struct {
	PageDrawer
	template *template.Template
	data     interface{}
}

func NewTemplateDrawer(tmpl string, data interface{}) *TemplateDrawer {
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		log.Print(err)
		return nil
	}
	return &TemplateDrawer{
		template: t,
		data:     data,
	}
}

func (ptr *TemplateDrawer) Draw(page *BufferedPage) {
	err := ptr.template.Execute(page, ptr.data)
	if err != nil {
		log.Print(err)
	}
}
