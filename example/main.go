package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"rverpi/ihui.v2"
)

// Button
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
	html := fmt.Sprintf(`<button id="%s">%s</button>`, b.id, b.label)
	page.WriteString(html)
	page.On(b.id, "click", b.action)
}

// Pages
func page1(page *ihui.Page) {
	page.WriteString(`<p>Hello page1</p>`)
	page.Draw(newButton("Exit", func(session *ihui.Session) {
		log.Println("close!")
	}))
}

func index(page *ihui.Page) {
	page.WriteString(`<p>Hello index</p>`)
	page.Draw(newButton("go page 1", func(session *ihui.Session) {
		p := session.NewPageFunc("Page 1", page1)
		session.ShowPage(p)
	}))
}

func index2(page *ihui.Page) {
	page.WriteString(`<p>Hello index 2</p>`)
	page.Draw(newButton("go page 1", func(session *ihui.Session) {
		p := session.NewPageFunc("Page 1", page1)
		session.ShowPage(p)
	}))
}

// Init
func start(session *ihui.Session) {
	menu := NewMenu()
	menu.Add("Page1", ihui.RenderFunc(index))
	menu.Add("Page2", ihui.RenderFunc(index2))

	page := session.NewPage("Example", menu)
	for {
		ev, err := session.ShowPage(page)
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

	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
