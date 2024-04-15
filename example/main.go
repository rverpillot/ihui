package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rverpillot/ihui"
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
	b.id = page.UniqueId("id-")
	fmt.Fprintf(page, `<button id="%s">%s</button>`, b.id, b.label)
	page.On("click", "[id="+b.id+"]", b.action)
}

// Pages
func modal1(page ihui.Page) {
	page.WriteString(`<p>Hello page1</p>`)
	button := newButton("Exit", func(session *ihui.Session, _ ihui.Event) {
		log.Println("close modal page!")
		session.CloseModalPage()
	})
	button.Render(page)

	page.On("create", "page", func(s *ihui.Session, _ ihui.Event) {
		log.Println("page1 loaded!")
	})
}

func tab1(page ihui.Page) {
	page.WriteString(`<p>Hello Tab 1</p>`)
}

func tab2(page ihui.Page) {
	page.WriteString(`<p>Hello Tab 2</p>`)
	button := newButton("go page 1", func(session *ihui.Session, event ihui.Event) {
		session.ShowPage("page1", ihui.PageRendererFunc(modal1), &ihui.Options{Title: "Page 1", Modal: true})
	})
	button.Render(page)
}

// Init
func start(session *ihui.Session) {
	menu := NewMenu()
	menu.Add("Tab1", ihui.PageRendererFunc(tab1))
	menu.Add("Tab2", ihui.PageRendererFunc(tab2))

	session.ShowPage("menu", menu, &ihui.Options{Title: "Example"})
}

func main() {
	http.Handle("/", ihui.NewHTTPHandler(start))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	addr := fmt.Sprintf("localhost:%s", port)
	log.Printf("Listen to http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
