
all: bin/example

resources/js/ihui.min.js: ihui.js 
	browserify ihui.js -o resources/js/ihui.js
	uglifyjs -c -o resources/js/ihui.min.js -- resources/js/ihui.js 

bin/example: example/*.go *.go resources/index.html ihui.js
	cd example && go build -o ../$@ .

clean: 
	rm resources/js/*.js bin/*


