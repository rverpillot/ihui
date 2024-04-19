package templating

import (
	"io"
	"io/fs"

	"github.com/cbroglie/mustache"
)

type PageMustache struct {
	fsys     fs.FS
	filename string
	template *mustache.Template
}

func NewPageMustache(fsys fs.FS, filename string) *PageMustache {
	return &PageMustache{
		fsys:     fsys,
		filename: filename,
	}
}

func (p *PageMustache) Reload() {
	p.template = nil
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
	return p.template.FRender(w, model)
}
