package main

import (
	"log"
	"net/http"
	"os"

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

func (b *Button) Render(page *ihui.Page) {
	b.id = page.NewId()
	html := fmt.Sprintf(`<button id="%s" data-action="click">%s</button>`, b.id, b.label)
	page.WriteString(html)
	page.On(b.id, "click", b.action)
}

func page1(page *ihui.Page) {
	page.WriteString(`<p>Hello page1</p>`)
	page.Add(newButton("Exit", func(session *ihui.Session) {
		log.Println("close!")
	}))
}

func index(page *ihui.Page) {
	page.WriteString(`<p>Hello index</p>`)
	page.Add(newButton("go page 1", func(session *ihui.Session) {
		p := ihui.NewPage(session, "Page 1", ihui.RenderFunc(page1))
		session.ShowPage(p)
	}))
}

func index2(page *ihui.Page) {
	page.WriteString(`<p>Hello index 2</p>`)
	page.Add(newButton("go page 1", func(session *ihui.Session) {
		p := ihui.NewPage(session, "Page 1", ihui.RenderFunc(page1))
		session.ShowPage(p)
	}))
}

func start(session *ihui.Session) {
	menu := NewMenu()
	menu.Add("Page1", ihui.RenderFunc(index))
	menu.Add("Page2", ihui.RenderFunc(index2))

	for {
		ev, err := session.ShowPage(ihui.NewPage(session, "Example", menu))
		if err != nil {
			break
		}
		log.Println(ev)
	}
}

func main() {

	h := ihui.NewHTTPHandler("/app", start)

	log.Println(h.Pattern())
	http.Handle(h.Pattern(), h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	log.Fatal(http.ListenAndServe("*:"+port, nil))
}
