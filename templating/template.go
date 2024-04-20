package templating

import (
	"html/template"
	"io"
	"io/fs"
)

type PageTemplate struct {
	pageName string
	content  string
	template *template.Template
}

func NewPageTemplate(pageName, content string) *PageTemplate {
	return &PageTemplate{
		pageName: pageName,
		content:  content,
	}
}

func (p *PageTemplate) Execute(w io.Writer, model interface{}) (err error) {
	if p.template == nil {
		p.template, err = template.New(p.pageName).Parse(p.content)
		if err != nil {
			return err
		}
	}
	return p.template.Execute(w, model)
}

type PageFileTemplate struct {
	fsys     fs.FS
	filename string
	template *template.Template
}

func NewPageFileTemplate(fsys fs.FS, filename string) *PageFileTemplate {
	return &PageFileTemplate{
		fsys:     fsys,
		filename: filename,
	}
}

func (p *PageFileTemplate) Reload() {
	p.template = nil
}

func (p *PageFileTemplate) Execute(w io.Writer, model interface{}) (err error) {
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
