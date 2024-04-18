package templating

import (
	"io"
	"io/fs"

	"github.com/cbroglie/mustache"
	"github.com/rverpillot/ihui"
)

type PageMustache struct {
	fsys     fs.FS
	filename string
	template *mustache.Template
	model    interface{}
}

func NewPageMustache(fsys fs.FS, filename string, model interface{}) *PageMustache {
	return &PageMustache{
		fsys:     fsys,
		filename: filename,
		model:    model,
	}
}

func (p *PageMustache) SetModel(model interface{}) {
	p.model = model
}

func (p *PageMustache) Execute(w io.Writer, model interface{}) (err error) {
	if p.template == nil {
		content, err := fs.ReadFile(p.fsys, p.filename) 
		if err != nil {
			return err
		}
		tpl, err := mustache.ParseString(string(content))
		if err != nil {
			return err
		}
		p.template = tpl
	}
	return p.template.FRender(w, p.model)
}

func (p *PageMustache) Render(page *ihui.Page) error {
	return p.Execute(page, p.model)
}
