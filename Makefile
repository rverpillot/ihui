
RESOURCES=$(find ihui_resources/ -type f)

all: bin/example

bindata_assetfs.go: $(RESOURCES) ihui.js
	browserify ihui.js -o resources/ihui/js/ihui.js
	go generate	

bin/example: example/*.go bindata_assetfs.go *.go
	go build -o $@ example/main.go example/menu.go


