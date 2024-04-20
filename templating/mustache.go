package templating

import (
	"io"
	"io/fs"

	"github.com/cbroglie/mustache"
)

type PageMustache struct {
	content  string
	template *mustache.Template
}

func NewPageMustache(content string) *PageMustache {
	return &PageMustache{
		content: content,
	}
}

func (p *PageMustache) Execute(w io.Writer, model interface{}) (err error) {
	if p.template == nil {
		p.template, err = mustache.ParseString(p.content)
		if err != nil {
			return
		}
	}
	return p.template.FRender(w, model)
}

type PageFileMustache struct {
	fsys     fs.FS
	filename string
	template *mustache.Template
}

func NewPageFileMustache(fsys fs.FS, filename string) *PageFileMustache {
	return &PageFileMustache{
		fsys:     fsys,
		filename: filename,
	}
}

func (p *PageFileMustache) Reload() {
	p.template = nil
}

func (p *PageFileMustache) Execute(w io.Writer, model interface{}) (err error) {
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
