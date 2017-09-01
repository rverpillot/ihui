mkdir -p src/rverpi/ihui.v2 && mv example *.go *.js vendor ihui_resources src/rverpi/ihui.v2/
export GOPATH=$(pwd)
APP=$(pwd)
go get github.com/jteeuwen/go-bindata/...
go get github.com/elazarl/go-bindata-assetfs/...
go get github.com/PuerkitoBio/goquery
go get github.com/gorilla/websocket
(cd src/rverpi/ihui.v2 && go generate)
go build -x -o bin/example src/rverpi/ihui.v2/example/main.go src/rverpi/ihui.v2/example/menu.go
