package ihui

import (
	"html/template"
	"log"
)

type PageTemplateDrawer struct {
	PageDrawer
	template *template.Template
	data     interface{}
}

func NewPageTemplateDrawer(tmpl string, data interface{}) *PageTemplateDrawer {
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		log.Print(err)
		return nil
	}
	return &PageTemplateDrawer{
		template: t,
		data:     data,
	}
}

func (ptr *PageTemplateDrawer) Draw(page *BufferedPage) {
	err := ptr.template.Execute(page, ptr.data)
	if err != nil {
		log.Print(err)
	}
}
