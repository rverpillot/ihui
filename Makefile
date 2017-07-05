
all: lib

ihui_resources/js/ihui.js: ihui.js
	browserify $< -o $@

rice-box.go: ihui_resources/js/* ihui_resources/*
	rice embed-go	

lib: rice-box.go
	go install -x