
all: bin/example

clean: 
	rm bindata.go resources/js/*.js bin/*

bindata.go: resources/index.html ihui.js
	browserify ihui.js -o resources/js/ihui.js
	uglifyjs -c -o resources/js/ihui.min.js -- resources/js/ihui.js 
	go generate	

bin/example: example/*.go bindata.go *.go
	cd example && go build -o ../$@ .


