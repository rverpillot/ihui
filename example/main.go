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

func (b *Button) Render(page ihui.Page) {
	b.id = page.UniqueId("id_")
	html := fmt.Sprintf(`<button id="%s">%s</button>`, b.id, b.label)
	page.WriteString(html)
	sel := "[id=" + b.id + "]"
	page.On("click", sel, func(session *ihui.Session, event ihui.Event) {
		log.Printf("click button! %s", event.Source)
	})
	page.On("click", sel, b.action)
}

// Pages
func page1(page ihui.Page) {
	page.WriteString(`<p>Hello page1</p>`)
	button := newButton("Exit", func(session *ihui.Session, _ ihui.Event) {
		log.Println("close!")
		session.QuitPage()
	})
	button.Render(page)

	page.On("load", "page", func(s *ihui.Session, _ ihui.Event) {
		log.Println("page1 loaded!")
	})
}

func tab1(page ihui.Page) {
	page.WriteString(`<p>Hello Tab 1</p>`)
}

func tab2(page ihui.Page) {
	page.WriteString(`<p>Hello Tab 2</p>`)
	button := newButton("go page 1", func(session *ihui.Session, event ihui.Event) {
		session.ShowPage(ihui.PageRendererFunc(page1), &ihui.Options{Title: "Page 1", Modal: true})
	})
	button.Render(page)
}

// Init
func start(session *ihui.Session) {
	menu := NewMenu()
	menu.Add("Tab1", ihui.PageRendererFunc(tab1))
	menu.Add("Tab2", ihui.PageRendererFunc(tab2))

	if err := session.ShowPage(menu, &ihui.Options{Title: "Example"}); err != nil {
		log.Print(err)
	}
}

func main() {

	h := ihui.NewHTTPHandler(start)

	http.Handle("/", h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
