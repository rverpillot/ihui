package templating

import (
	"html/template"
	"io"
	"io/fs"

	"github.com/yosssi/ace"
)

type PageFileAce struct {
	fsys     fs.FS
	filename string
	template *template.Template
}

func NewPageFileAce(fsys fs.FS, filename string) *PageFileAce {
	return &PageFileAce{
		fsys:     fsys,
		filename: filename,
	}
}

func (p *PageFileAce) Reload() {
	p.template = nil
}

func (p *PageFileAce) Execute(w io.Writer, model interface{}) error {
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
