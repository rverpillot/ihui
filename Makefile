
RESOURCES=$(find ihui_resources/ -type f)

all: lib bin/example

bindata_assetfs.go: $(RESOURCES) ihui.js
	browserify $(RESOURCES) -o ihui_resources/js/ihui.js
	go generate	

lib: *.go
	go install -v

bin/example: example/main.go example/menu.go
	go build -o $@ $?

deps:
	gvt fetch github.com/jteeuwen/go-bindata
	gvt fetch github.com/elazarl/go-bindata-assetfs
	gvt fetch github.com/PuerkitoBio/goquery
	gvt fetch github.com/gorilla/websocket