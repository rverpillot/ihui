package templating

import (
	"html/template"
	"io"
	"io/fs"

	"github.com/rverpillot/ihui"
	"github.com/yosssi/ace"
)

type PageAce struct {
	fsys     fs.FS
	filename string
	template *template.Template
	model    interface{}
}

func NewPageAce(fsys fs.FS, filename string, model interface{}) *PageAce {
	return &PageAce{
		fsys:     fsys,
		filename: filename,
		model:    model,
	}
}

func (p *PageAce) SetModel(model interface{}) {
	p.model = model
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

func (p *PageAce) Render(page *ihui.Page) error {
	return p.Execute(page, p.model)
}
