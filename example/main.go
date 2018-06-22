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

func (b *Button) Draw(page ihui.Page) {
	b.id = page.NewId()
	html := fmt.Sprintf(`<button id="%s">%s</button>`, b.id, b.label)
	page.WriteString(html)
	page.On(b.id, "click", b.action)
}

// Pages
func page1(page ihui.Page) {
	page.WriteString(`<p>Hello page1</p>`)
	page.Draw(newButton("Exit", func(session *ihui.Session) {
		log.Println("close!")
	}))
}

func tab1(page ihui.Page) {
	page.WriteString(`<p>Hello Tab 1</p>`)
}

func tab2(page ihui.Page) {
	page.WriteString(`<p>Hello Tab 2</p>`)
	page.Draw(newButton("go page 1", func(session *ihui.Session) {
		session.ShowPage("Page 1", ihui.PageDrawerFunc(page1))
	}))
}

// Init
func start(session *ihui.Session) {
	menu := NewMenu()
	menu.Add("Tab1", ihui.PageDrawerFunc(tab1))
	menu.Add("Tab2", ihui.PageDrawerFunc(tab2))

	for {
		if err := session.ShowPage("Exemple", menu); err != nil {
			log.Print(err)
			break
		}
	}
}

func main() {

	h := ihui.NewHTTPHandler("/app", start)

	http.Handle("/", h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
