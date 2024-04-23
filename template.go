package ihui

import (
	"html/template"
	"io"
	"io/fs"
	"path"
)

type GoTemplate struct {
	pageName string
	content  string
	template *template.Template
}

func NewGoTemplate(pageName, content string) *GoTemplate {
	return &GoTemplate{
		pageName: pageName,
		content:  content,
	}
}

func (p *GoTemplate) Execute(w io.Writer, model interface{}) (err error) {
	if p.template == nil {
		p.template, err = template.New(p.pageName).Parse(p.content)
		if err != nil {
			return err
		}
	}
	return p.template.Execute(w, model)
}

type GoTemplateFile struct {
	fsys     fs.FS
	filename string
	template *template.Template
}

func NewGoTemplateFile(fsys fs.FS, filename string) *GoTemplateFile {
	return &GoTemplateFile{
		fsys:     fsys,
		filename: filename,
	}
}

func (p *GoTemplateFile) Reload() {
	p.template = nil
}

func (p *GoTemplateFile) Execute(w io.Writer, model interface{}) (err error) {
	if p.template == nil {
		p.template, err = template.New(path.Base(p.filename)).ParseFS(p.fsys, p.filename)
		if err != nil {
			return err
		}
	}
	return p.template.Execute(w, model)
}
