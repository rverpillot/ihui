
all: lib

ihui_resources/js/ihui.js: ihui.js
	browserify $< -o $@

lib: 
	go generate
	go install -x