package templating

import (
	"io"
	"io/fs"

	"github.com/cbroglie/mustache"
)

type MustacheTemplate struct {
	content  string
	template *mustache.Template
}

func NewMustacheTemplate(content string) *MustacheTemplate {
	return &MustacheTemplate{
		content: content,
	}
}

func (p *MustacheTemplate) Execute(w io.Writer, model interface{}) (err error) {
	if p.template == nil {
		p.template, err = mustache.ParseString(p.content)
		if err != nil {
			return
		}
	}
	return p.template.FRender(w, model)
}

type MustacheTemplateFile struct {
	fsys     fs.FS
	filename string
	template *mustache.Template
}

func NewMustacheTemplateFile(fsys fs.FS, filename string) *MustacheTemplateFile {
	return &MustacheTemplateFile{
		fsys:     fsys,
		filename: filename,
	}
}

func (p *MustacheTemplateFile) Reload() {
	p.template = nil
}

func (p *MustacheTemplateFile) Execute(w io.Writer, model interface{}) (err error) {
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
	return p.template.FRender(w, model)
}
