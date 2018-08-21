package ihui

import (
	"html/template"
	"log"

	"github.com/yosssi/ace"
)

type PageAceTemplateDrawer struct {
	PageDrawer
	template *template.Template
	data     interface{}
}

func NewPageAceTemplateDrawer(content []byte, data interface{}) *PageAceTemplateDrawer {
	options := ace.InitializeOptions(nil)
	options.Asset = func(name string) ([]byte, error) {
		return content, nil
	}

	t, err := ace.Load("content", "", options)
	if err != nil {
		panic(err)
	}
	return &PageAceTemplateDrawer{
		template: t,
		data:     data,
	}
}

func (ptr *PageAceTemplateDrawer) Draw(page *BufferedPage) {
	err := ptr.template.Execute(page, ptr.data)
	if err != nil {
		log.Print(err)
	}
}
