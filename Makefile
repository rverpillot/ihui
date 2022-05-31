
all: bin/example

clean: 
	rm resources/js/*.js bin/*

bin/example: example/*.go *.go resources/index.html ihui.js
	browserify ihui.js -o resources/js/ihui.js
	uglifyjs -c -o resources/js/ihui.min.js -- resources/js/ihui.js 
	cd example && go build -o ../$@ .


