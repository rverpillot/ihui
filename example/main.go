package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/rverpillot/ihui"
)

//go:embed statics
var StaticsFs embed.FS

// Pages
func modal1(e *ihui.HTMLElement) error {
	e.WriteString(`
	<section class="section box">
		<div class="block">
			<p>Hello page modal</p>
		</div>
		<div class="field is-grouped">
			<button class="button is-primary is-small" data-id="exit">Exit</button>
			<button class="button is-danger is-small" data-id="error">Error</button>
			<button class="button is-danger is-small" data-id="quit">Quit Application</button>
		</div>
	</section>
	`)
	e.On("click", "[data-id=exit]", func(s *ihui.Session, _ ihui.Event) error {
		e.Close()
		return nil
	})
	e.On("click", "[data-id=error]", func(s *ihui.Session, _ ihui.Event) error {
		return fmt.Errorf("an error occured")
	})
	e.On("click", "[data-id=quit]", func(s *ihui.Session, _ ihui.Event) error {
		return s.Quit()
	})
	e.On("element-created", "", func(s *ihui.Session, _ ihui.Event) error {
		log.Println("page1 loaded!")
		return nil
	})
	return nil
}

func tab1(e *ihui.HTMLElement) error {
	e.WriteString(`
	<div class="block">
		<p>Hello Tab 1</p>
	</div>
	`)
	return nil
}

func tab2(e *ihui.HTMLElement) error {
	e.WriteString(`
	<div class="block">
		<p>Hello Tab 2</p>
	</div>
	<button class="button is-link is-small" data-id="modal1">Modal 1</button>
	`)
	e.On("click", "[data-id=modal1]", func(s *ihui.Session, _ ihui.Event) error {
		return s.ShowModal("modal1", ihui.HTMLRendererFunc(modal1), ihui.Options{Title: "Modal 1"})
	})
	return nil
}

// Init
func start(session *ihui.Session) error {
	menu := NewMenu()
	menu.AddItem("Tab1", ihui.HTMLRendererFunc(tab1))
	menu.AddItem("Tab2", ihui.HTMLRendererFunc(tab2))
	return session.ShowPage("menu", menu, ihui.Options{Title: "Example"})
}

func main() {
	fsys, _ := fs.Sub(StaticsFs, "statics")
	http.Handle("/", http.FileServer(http.FS(fsys)))
	ihui.Handle("/", start)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	addr := fmt.Sprintf("localhost:%s", port)
	log.Printf("Listen to http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
