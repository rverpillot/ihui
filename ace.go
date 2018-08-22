package ihui

import (
	"html/template"
	"log"

	"github.com/yosssi/ace"
)

type PageAce struct {
	template *template.Template
	data     interface{}
}

func NewPageAce(content []byte, data interface{}) *PageAce {
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
		data:     data,
	}
}

func (ptr *PageAce) Render(page Page) {
	err := ptr.template.Execute(page, ptr.data)
	if err != nil {
		log.Print(err)
	}
}
