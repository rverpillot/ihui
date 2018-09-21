package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bitbucket.org/rverpi90/ihui"
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
	html := fmt.Sprintf(`<button id="%s">%s</button>`, b.id, b.label)
	page.WriteString(html)
	sel := "[id=" + b.id + "]"
	page.On("click", sel, func(session *ihui.Session, event ihui.Event) bool {
		log.Printf("click button! %s", event.Id)
		return false
	})
	page.On("click", sel, b.action)
}

// Pages
func page1(page ihui.Page) {
	page.WriteString(`<p>Hello page1</p>`)
	button := newButton("Exit", func(session *ihui.Session, _ ihui.Event) bool {
		log.Println("close!")
		return session.CloseModalPage()
	})
	button.Render(page)

	page.On("create", "page", func(s *ihui.Session, _ ihui.Event) bool {
		log.Println("page1 loaded!")
		return false
	})
}

func tab1(page ihui.Page) {
	page.WriteString(`<p>Hello Tab 1</p>`)
}

func tab2(page ihui.Page) {
	page.WriteString(`<p>Hello Tab 2</p>`)
	button := newButton("go page 1", func(session *ihui.Session, event ihui.Event) bool {
		return session.ShowPage("page1", ihui.PageRendererFunc(page1), &ihui.Options{Title: "Page 1", Modal: true})
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

	h := ihui.NewHTTPHandler(start)

	http.Handle("/", h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Listen to http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
