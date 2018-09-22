
all: bin/example

bindata.go: resources/index.html ihui.js
	browserify ihui.js -o resources/js/ihui.js
	uglify -s resources/js/ihui.js -o resources/js/ihui.min.js
	go generate	

bin/example: example/*.go bindata.go *.go
	cd example && go build -o ../$@ .


