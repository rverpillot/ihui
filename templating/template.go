package templating

import (
	"html/template"
	"io"
	"io/fs"
)

type PageTemplate struct {
	fsys     fs.FS
	filename string
	template *template.Template
}

func NewPageTemplate(fsys fs.FS, filename string) *PageTemplate {
	return &PageTemplate{
		fsys:     fsys,
		filename: filename,
	}
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
