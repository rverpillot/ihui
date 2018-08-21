package ihui

import (
	"html/template"
	"log"

	"github.com/yosssi/ace"
)

type AceTemplateDrawer struct {
	PageDrawer
	template *template.Template
	data     interface{}
}

func NewAceTemplateDrawer(content []byte, data interface{}) *AceTemplateDrawer {
	options := ace.InitializeOptions(nil)
	options.Asset = func(name string) ([]byte, error) {
		return content, nil
	}

	t, err := ace.Load("content", "", options)
	if err != nil {
		panic(err)
	}
	return &AceTemplateDrawer{
		template: t,
		data:     data,
	}
}

func (ptr *AceTemplateDrawer) Draw(page *BufferedPage) {
	err := ptr.template.Execute(page, ptr.data)
	if err != nil {
		log.Print(err)
	}
}
