mkdir -p src/rverpi/ihui.v2 && mv example *.go *.js vendor ihui_resources src/rverpi/ihui.v2/
ls -R /var/app
export GOPATH=$(pwd)
go get github.com/GeertJohan/go.rice
go get github.com/PuerkitoBio/goquery
go get github.com/gorilla/websocket
go build -x -o bin/example src/rverpi/ihui.v2/example/main.go src/rverpi/ihui.v2/example/menu.go
exit $?