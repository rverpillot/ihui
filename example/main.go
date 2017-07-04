package main

import (
	"log"
	"net/http"

	"fmt"

	"rverpi/ihui.v2"
)

type Button struct {
	id    string
	label string
}

func drawButton(page *ihui.Page, label string) string {
	id := page.NewId()
	html := fmt.Sprintf(`<button id="%s" data-action="click">%s</button>`, id, label)
	page.WriteString(html)

	page.On(id, "click", func(ctx *ihui.Context) {
		log.Println("Click button!")
	})
	return id
}

func (b *Button) Draw(page *ihui.Page) {
	b.id = page.NewId()
	html := fmt.Sprintf(`<button id="%s" data-action="click">%s</button>`, b.id, b.label)
	page.WriteString(html)

	page.On(b.id, "click", func(ctx *ihui.Context) {
		log.Println("Click button!")
	})
}

func page1(page *ihui.Page) {
	page.WriteString(`<p>Hello page1</p>`)
	buttonID := drawButton(page, "Exit")

	page.On(buttonID, "click", func(ctx *ihui.Context) {
		log.Println("close!")
	})
}

func index(page *ihui.Page) {
	page.WriteString(`<p>Hello index</p>`)
	buttonID := drawButton(page, "go page 1")

	page.On(buttonID, "click", func(ctx *ihui.Context) {
		log.Println(ctx.NewPage("page1", "Page 1", ihui.RenderFunc(page1)).Show(false))
	})

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
