package templating

import (
	"html/template"
	"io"
	"io/fs"

	"github.com/rverpillot/ihui"
)

type PageTemplate struct {
	fsys     fs.FS
	filename string
	template *template.Template
	model    interface{}
}

func NewPageTemplate(fsys fs.FS, filename string, model interface{}) *PageTemplate {
	return &PageTemplate{
		fsys:     fsys,
		filename: filename,
		model:    model,
	}
}

func (p *PageTemplate) SetModel(model interface{}) {
	p.model = model
}

func (p *PageTemplate) Reload() {
	p.template = nil
}

func (p *PageTemplate) Execute(w io.Writer, model interface{}) (err error) {
	if p.template == nil {
		content, err := fs.ReadFile(p.fsys, p.filename)
		if err != nil {
			return err
		}
		p.template, err = template.New(p.filename).Parse(string(content))
		if err != nil {
			return err
		}
	}
	return p.template.Execute(w, model)
}

func (p *PageTemplate) Render(page *ihui.Page) error {
	return p.Execute(page, p.model)
}
