
RESOURCES=$(find ihui_resources/ -type f)

all: lib bin/example

bindata_assetfs.go: $(RESOURCES) ihui.js
	browserify ihui.js -o ihui_resources/js/ihui.js
	go generate	

lib: *.go bindata_assetfs.go
	go install -v

bin/example: example/*.go *.go
	go build -o $@ $?

deps:
	govend -sv

