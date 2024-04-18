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
	action ihui.ActionCallback
}

func newButton(label string, action ihui.ActionCallback) *Button {
	return &Button{
		label:  label,
		action: action,
	}
}

func (b *Button) Render(page *ihui.Page) {
	b.id = page.UniqueId("id-")
	fmt.Fprintf(page, `<button id="%s">%s</button>`, b.id, b.label)
	page.On("click", "[id="+b.id+"]", b.action)
}

// Pages
func modal1(page *ihui.Page) error {
	page.WriteString(`<p>Hello page1</p>`)
	button := newButton("Exit", func(session *ihui.Session, _ ihui.Event) error {
		page.Close()
		return nil
	})
	button.Render(page)
	button2 := newButton("Error", func(session *ihui.Session, _ ihui.Event) error {
		return fmt.Errorf("an error occured")
	})
	button2.Render(page)

	page.On("page-created", "", func(s *ihui.Session, _ ihui.Event) error {
		log.Println("page1 loaded!")
		return nil
	})
	return nil
}

func tab1(page *ihui.Page) error {
	page.WriteString(`<p>Hello Tab 1</p>`)
	return nil
}

func tab2(page *ihui.Page) error {
	page.WriteString(`<p>Hello Tab 2</p>`)
	button := newButton("go page 1", func(session *ihui.Session, event ihui.Event) error {
		return session.ShowPage("modal1", ihui.HTMLRendererFunc(modal1), &ihui.Options{Title: "Modal 1", Modal: true})
	})
	button.Render(page)
	return nil
}

// Init
func start(session *ihui.Session) error {
	menu := NewMenu()
	menu.Add("Tab1", ihui.HTMLRendererFunc(tab1))
	menu.Add("Tab2", ihui.HTMLRendererFunc(tab2))

	return session.ShowPage("menu", menu, &ihui.Options{Title: "Example"})
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
