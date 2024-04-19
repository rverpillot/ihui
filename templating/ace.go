package templating

import (
	"html/template"
	"io"
	"io/fs"

	"github.com/yosssi/ace"
)

type PageAce struct {
	fsys     fs.FS
	filename string
	template *template.Template
}

func NewPageAce(fsys fs.FS, filename string) *PageAce {
	return &PageAce{
		fsys:     fsys,
		filename: filename,
	}
}

func (p *PageAce) Reload() {
	p.template = nil
}

func (p *PageAce) Execute(w io.Writer, model interface{}) error {
	if p.template == nil {
		content, err := fs.ReadFile(p.fsys, p.filename)
		if err != nil {
			return err
		}
		options := &ace.Options{
			Asset: func(name string) ([]byte, error) {
				return content, nil
			},
		}
		tpl, err := ace.Load(p.filename, "", options)
		if err != nil {
			return err
		}
		p.template = tpl
	}
	return p.template.Execute(w, model)
}
