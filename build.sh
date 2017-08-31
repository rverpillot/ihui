mkdir rverpi/ihui.v2 && mv example *.go *.js node_modules vendor ihui_resources rverpi/ihui.v2/
go build -x -o bin/example rverpi/ihui.v2/example/main.go rverpi/ihui.v2/example/menu.go
