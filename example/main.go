package main

import (
	"log"
	"net/http"

	"fmt"

	"rverpi/ihui.v2"
)

type Button struct {
	id     string
	label  string
	action ihui.ActionFunc
}

func newButton(label string, action ihui.ActionFunc) *Button {
	return &Button{
		label:  label,
		action: action,
	}
}

func (b *Button) Draw(page *ihui.Page) {
	b.id = page.NewId()
	html := fmt.Sprintf(`<button id="%s" data-action="click">%s</button>`, b.id, b.label)
	page.WriteString(html)
	page.On(b.id, "click", b.action)
}

func page1(page *ihui.Page) {
	page.WriteString(`<p>Hello page1</p>`)
	page.Add(newButton("Exit", func(ctx *ihui.Context) {
		log.Println("close!")
	}))
}

func index(page *ihui.Page) {
	page.WriteString(`<p>Hello index</p>`)
	page.Add(newButton("go page 1", func(ctx *ihui.Context) {
		ctx.NewPage("page1", "Page 1", ihui.RenderFunc(page1)).Show(false)
	}))
}

func start(ctx *ihui.Context) {
	index := ctx.NewPage("index", "Hello", ihui.RenderFunc(index))

	for {
		ev, err := index.Show(false)
		if err != nil {
			break
		}
		log.Println(ev)
	}
}

func main() {

	h := ihui.NewHTTPHandler("Sample", start)

	http.Handle("/app/", h)

	log.Fatal(http.ListenAndServe("localhost:9090", nil))
}
