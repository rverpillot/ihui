
RESOURCES=$(find ihui_resources/ -type f)

all: bin/example

bindata.go: $(RESOURCES) ihui.js
	browserify ihui.js -o resources/js/ihui.js
	vgo generate	

bin/example: example/*.go bindata.go *.go
	cd example && vgo build -o ../$@ .


