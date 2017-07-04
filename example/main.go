package main

import (
	"io"
	"log"
	"net/http"

	"rverpi/ihui.v2"
)

func page1(w io.Writer) {
	w.Write([]byte(`<p>Hello page1</p>`))
	w.Write([]byte(`<a id="close" href="#" data-action="click">Close!</a>`))
}

func index(w io.Writer) {
	w.Write([]byte(`<p>Hello index</p>`))
	w.Write([]byte(`<a id="go" href="#" data-action="click">Action!</a>`))
}

func start(ctx *ihui.Context) {
	index := ihui.NewPage("index", "Hello", ihui.RenderFunc(index))
	page := ihui.NewPage("page1", "Page 1", ihui.RenderFunc(page1))

	index.On("go.click", func(ctx *ihui.Context) {
		log.Println(ctx.DisplayPage(page, false))
	})

	for {
		ev, err := ctx.DisplayPage(index, false)
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
